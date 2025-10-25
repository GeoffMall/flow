package parser

import (
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLDecoder is a thin wrapper over yaml.Decoder that:
//  1. Decodes a YAML stream document-by-document,
//  2. Normalizes each document to JSON-compatible Go types
//     (map[string]any, []any, string, bool, float64, etc.)
type YAMLDecoder struct {
	dec *yaml.Decoder
}

// NewYAMLDecoder creates a streaming YAML decoder.
// It does not read the entire input into memory.
func NewYAMLDecoder(r io.Reader) *YAMLDecoder {
	return &YAMLDecoder{dec: yaml.NewDecoder(r)}
}

// Next decodes the next YAML document from the stream.
// It returns (value, ok=false, nil) on EOF, or (nil, false, err) on an error.
// On success, 'ok' is true and 'value' is already normalized to JSON-compatible types.
func (d *YAMLDecoder) Next() (any, bool, error) {
	var node any
	if err := d.dec.Decode(&node); err != nil {
		if err == io.EOF {
			return nil, false, nil
		}
		return nil, false, err
	}
	return normalizeYAML(node), true, nil
}

// YAMLEncoder is a small helper around yaml.Encoder.
// It streams output and avoids buffering the entire structure.
type YAMLEncoder struct {
	enc *yaml.Encoder
}

// NewYAMLEncoder creates an encoder with a modest default indent (2).
// Call Close() when done to flush the stream.
func NewYAMLEncoder(w io.Writer) *YAMLEncoder {
	e := yaml.NewEncoder(w)
	e.SetIndent(2)
	return &YAMLEncoder{enc: e}
}

// Encode writes a single YAML document (with trailing newline).
func (e *YAMLEncoder) Encode(v any) error {
	// We assume v is already JSON-compatible (map[string]any, []any, etc.),
	// which yaml.v3 handles just fine.
	return e.enc.Encode(v)
}

// Close flushes the encoder and frees resources.
// Always call this when you're done encoding.
func (e *YAMLEncoder) Close() error {
	return e.enc.Close()
}
