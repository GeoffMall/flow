package yaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser implements format.Parser for YAML format.
// It streams YAML documents separated by --- markers.
type Parser struct {
	dec *yaml.Decoder
}

// NewParser creates a new YAML streaming parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		dec: yaml.NewDecoder(r),
	}
}

// ForEach streams YAML documents and calls fn for each.
// Documents are separated by --- markers in YAML.
// All values are normalized to JSON-compatible Go types.
func (p *Parser) ForEach(fn func(any) error) error {
	for {
		var node any
		if err := p.dec.Decode(&node); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		// Normalize YAML types to JSON-compatible types
		normalized := normalizeYAML(node)

		if err := fn(normalized); err != nil {
			return err
		}
	}
}

// normalizeYAML converts yaml.v3 decoded values into JSON-compatible Go types:
//   - map[any]any  -> map[string]any (recursively)
//   - []any        -> []any (recursively)
//   - scalar nodes -> left as-is
//
// This ensures operations work consistently across JSON and YAML input.
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

// toStringKey converts any key type to a string for map keys.
// This handles YAML's more flexible key types (numbers, booleans, etc.).
func toStringKey(k any) string {
	switch t := k.(type) {
	case string:
		return t
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return fmt.Sprint(t)
	case json.Number:
		return t.String()
	case []byte:
		return string(t)
	default:
		// Fallback: use fmt.Sprint
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
