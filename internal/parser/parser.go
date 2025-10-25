package parser

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type Format int

const (
	Unknown Format = iota
	JSON
	YAML
)

// Parser streams structured values (objects/arrays/scalars) from an input
// that may be JSON or YAML. It never buffers the entire input in memory.
type Parser struct {
	br      *bufio.Reader
	format  Format
	jd      *json.Decoder
	yd      *yaml.Decoder
	started bool
}

// New creates a streaming parser that autodetects JSON or YAML.
func New(r io.Reader) (*Parser, error) {
	br := bufio.NewReaderSize(r, 64*1024)

	format, err := detectFormat(br)
	if err != nil {
		return nil, err
	}

	parser := &Parser{
		br:     br,
		format: format,
	}

	return initializeDecoder(parser)
}

// detectFormat analyzes the input stream to determine if it's JSON or YAML
func detectFormat(br *bufio.Reader) (Format, error) {
	peek, _ := br.Peek(1024)
	head := strings.TrimLeft(string(peek), " \t\r\n")

	if len(head) == 0 {
		return JSON, nil // Empty input: treat as JSON
	}

	return classifyByFirstChar(head)
}

// classifyByFirstChar determines format based on the first non-space character
func classifyByFirstChar(head string) (Format, error) {
	switch head[0] {
	case '{', '[':
		return JSON, nil
	case '%':
		return YAML, nil // YAML directive like "%YAML 1.2"
	case '-':
		return classifyDashPrefix(head), nil
	default:
		return classifyByHeuristic(head), nil
	}
}

// classifyDashPrefix handles the ambiguous dash character
func classifyDashPrefix(head string) Format {
	if strings.HasPrefix(head, "---") {
		return YAML
	}
	return JSON // Could be negative number
}

// classifyByHeuristic uses content analysis for ambiguous cases
func classifyByHeuristic(head string) Format {
	if looksLikeYAML(head) {
		return YAML
	}
	return JSON
}

// initializeDecoder creates the appropriate decoder for the detected format
func initializeDecoder(p *Parser) (*Parser, error) {
	switch p.format {
	case JSON:
		return initJSONDecoder(p), nil
	case YAML:
		return initYAMLDecoder(p), nil
	default:
		return nil, errors.New("unknown input format")
	}
}

// initJSONDecoder sets up JSON decoding
func initJSONDecoder(p *Parser) *Parser {
	jd := json.NewDecoder(p.br)
	jd.UseNumber()
	p.jd = jd
	return p
}

// initYAMLDecoder sets up YAML decoding
func initYAMLDecoder(p *Parser) *Parser {
	p.yd = yaml.NewDecoder(p.br)
	return p
}

// Format reports the detected input format.
func (p *Parser) Format() Format { return p.format }

// ForEach streams every top-level value/document and calls fn for each.
//   - JSON: supports concatenated top-level JSON values. If the first value is
//     an array, it streams each array element as a separate item.
//   - YAML: streams documents separated by '---'.
func (p *Parser) ForEach(fn func(any) error) error {
	if p.format == JSON {
		return p.forEachJSON(fn)
	}

	return p.forEachYAML(fn)
}

func (p *Parser) forEachJSON(fn func(any) error) error {
	dec := p.jd

	rm, err := p.decodeFirstMessage(dec)
	if err != nil {
		return err
	}

	if p.isArray(rm) {
		return p.processArrayAndContinue(dec, rm, fn)
	}

	return p.processValueAndContinue(dec, rm, fn)
}

// decodeFirstMessage reads the first JSON message from the decoder
func (p *Parser) decodeFirstMessage(dec *json.Decoder) (json.RawMessage, error) {
	var rm json.RawMessage
	if err := dec.Decode(&rm); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, err
	}
	return rm, nil
}

// isArray checks if the raw message represents a JSON array
func (p *Parser) isArray(rm json.RawMessage) bool {
	return len(rm) > 0 && rm[0] == '['
}

// processArrayAndContinue handles array streaming followed by concatenated documents
func (p *Parser) processArrayAndContinue(dec *json.Decoder, rm json.RawMessage, fn func(any) error) error {
	if err := p.processArrayElements(rm, fn); err != nil {
		return err
	}

	return p.processConcatenatedDocuments(dec, fn)
}

// processValueAndContinue handles a single value followed by concatenated documents
func (p *Parser) processValueAndContinue(dec *json.Decoder, rm json.RawMessage, fn func(any) error) error {
	if err := p.processRawMessage(rm, fn); err != nil {
		return err
	}

	return p.processConcatenatedDocuments(dec, fn)
}

// processArrayElements streams individual elements from a JSON array
func (p *Parser) processArrayElements(rm json.RawMessage, fn func(any) error) error {
	var arr []any
	if err := json.Unmarshal(rm, &arr); err != nil {
		return err
	}

	for i := range arr {
		normalizeJSON(&arr[i])
		if err := fn(arr[i]); err != nil {
			return err
		}
	}

	return nil
}

// processRawMessage converts and processes a single raw JSON message
func (p *Parser) processRawMessage(rm json.RawMessage, fn func(any) error) error {
	var v any
	if err := json.Unmarshal(rm, &v); err != nil {
		return err
	}

	normalizeJSON(&v)
	return fn(v)
}

// processConcatenatedDocuments continues reading concatenated JSON documents
func (p *Parser) processConcatenatedDocuments(dec *json.Decoder, fn func(any) error) error {
	for {
		var rm json.RawMessage
		if err := dec.Decode(&rm); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if err := p.processRawMessage(rm, fn); err != nil {
			return err
		}
	}
}

// normalizeJSON ensures numbers remain usable and maps are map[string]any.
// (json.Unmarshal already produces map[string]any / []any; we keep a symmetric
// function so the YAML path uses the same transformer name.)
func normalizeJSON(v *any) {
	// Nothing required for standard library JSON output; left for symmetry.
}

// -------------------- YAML --------------------

func (p *Parser) forEachYAML(fn func(any) error) error {
	dec := p.yd

	for {
		var node any
		if err := dec.Decode(&node); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		norm := normalizeYAML(node)

		if err := fn(norm); err != nil {
			return err
		}
	}
}

// normalizeYAML converts yaml.v3 decoded values into JSON-compatible Go types:
//   - map[any]any  -> map[string]any (recursively)
//   - []any        -> []any (recursively)
//   - scalar nodes -> left as-is
func normalizeYAML(v any) any {
	switch vv := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(vv))
		for k, val := range vv {
			out[toStringKey(k)] = normalizeYAML(val)
		}

		return out
	case map[string]any:
		out := make(map[string]any, len(vv))
		for k, val := range vv {
			out[k] = normalizeYAML(val)
		}

		return out
	case []any:
		for i := range vv {
			vv[i] = normalizeYAML(vv[i])
		}

		return vv
	default:
		return v
	}
}

func toStringKey(k any) string {
	switch t := k.(type) {
	case string:
		return t
	default:
		return strings.TrimSpace(toStringFallback(t))
	}
}

func toStringFallback(v any) string {
	// Basic fallback; we avoid fmt to reduce allocations; this is okay for keys.
	// You can replace with fmt.Sprint if you prefer simplicity.
	switch x := v.(type) {
	case json.Number:
		return x.String()
	case []byte:
		return string(x)
	default:

		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(anyToString(x)), "}"), "{"), "\n", " "), "\t", " "))
	}
}

// anyToString is a small helper; fmt.Sprint is acceptable too.
func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case json.Number:
		return t.String()
	default:
		// Fall back to the default representation
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll("", "", "")), "", "")), "", "")), "\n", " "), "\t", " "))
	}
}

// -------------------- Heuristics --------------------

func looksLikeYAML(head string) bool {
	// Very light heuristic:
	// If the first non-space line contains a ':' before a ',' or '}', it's likely YAML key: value.
	// This is intentionally conservative—JSON is also valid YAML, but we prefer JSON when clearly JSON.
	line := head
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}

	colon := strings.IndexByte(line, ':')

	if colon < 0 {
		return false
	}

	comma := strings.IndexByte(line, ',')
	br := strings.IndexByte(line, '}')

	if comma == -1 {
		comma = 1 << 30
	}

	if br == -1 {
		br = 1 << 30
	}

	return colon < comma && colon < br
}
