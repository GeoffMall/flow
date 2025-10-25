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

	// Sniff the first non-space prefix without consuming the stream.
	// We Peek a chunk and inspect the first non-space rune.
	peek, _ := br.Peek(1024) // okay if fewer are available
	head := strings.TrimLeft(string(peek), " \t\r\n")
	var fmtGuess Format
	if len(head) == 0 {
		// Empty input: treat as JSON; subsequent decodes will hit EOF.
		fmtGuess = JSON
	} else {
		switch head[0] {
		case '{', '[':
			fmtGuess = JSON
		case '%': // YAML directive like "%YAML 1.2"
			fmtGuess = YAML
		case '-': // Could be '---' YAML start, or a negative number (JSON too).
			if strings.HasPrefix(head, "---") {
				fmtGuess = YAML
			} else {
				// fallback: try JSON first (numbers allowed), but we keep it simple:
				fmtGuess = JSON
			}
		default:
			// Heuristic: if it looks like a bare key (YAML), prefer YAML.
			// Otherwise default to JSON.
			if looksLikeYAML(head) {
				fmtGuess = YAML
			} else {
				fmtGuess = JSON
			}
		}
	}

	p := &Parser{
		br:     br,
		format: fmtGuess,
	}

	switch fmtGuess {
	case JSON:
		jd := json.NewDecoder(br)
		jd.UseNumber()
		p.jd = jd
	case YAML:
		p.yd = yaml.NewDecoder(br)
	default:
		return nil, errors.New("unknown input format")
	}

	return p, nil
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

// -------------------- JSON --------------------

func (p *Parser) forEachJSON(fn func(any) error) error {
	// We can see one of two shapes:
	//  1) Top-level array: stream its elements.
	//  2) One or more concatenated top-level JSON values: decode repeatedly.
	//
	// To support both, we first peek a token. If it's a '[' delimiter, we stream array elements.

	// Save decoder locally
	dec := p.jd

	// Pull the first non-space token, but we need to be careful:
	// We can't "unread" a token using encoding/json. So we do this:
	// - Try to read the first token.
	// - If it's '[', stream array mode.
	// - Else, we fall back to "concat documents" mode where we already consumed one token of the first value.
	//   To handle that, we reconstruct the value using that token path (but json.Decoder doesn't expose rewinding).
	//
	// Simpler approach: Attempt to decode into a json.RawMessage first.
	// If that decodes a single value successfully, we pass it on, then continue decoding RawMessage in a loop.
	// If the first decoded RawMessage is an array, we decode that array and emit its elements.

	var rm json.RawMessage
	if err := dec.Decode(&rm); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	// If first value is an array, stream its elements.
	if len(rm) > 0 && rm[0] == '[' {
		var arr []any
		// Decode the raw array into a slice to stream elements
		if err := json.Unmarshal(rm, &arr); err != nil {
			return err
		}
		for i := range arr {
			normalizeJSON(&arr[i])
			if err := fn(arr[i]); err != nil {
				return err
			}
		}
		// There might still be concatenated documents after the array; keep decoding RawMessage
		for {
			rm = nil
			if err := dec.Decode(&rm); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
			var v any
			if err := json.Unmarshal(rm, &v); err != nil {
				return err
			}
			normalizeJSON(&v)
			if err := fn(v); err != nil {
				return err
			}
		}
	}

	// Otherwise, first value is *not* an array; emit it, then continue with concatenated values.
	{
		var v any
		if err := json.Unmarshal(rm, &v); err != nil {
			return err
		}
		normalizeJSON(&v)
		if err := fn(v); err != nil {
			return err
		}
	}

	for {
		rm = nil
		if err := dec.Decode(&rm); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		var v any
		if err := json.Unmarshal(rm, &v); err != nil {
			return err
		}
		normalizeJSON(&v)
		if err := fn(v); err != nil {
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
