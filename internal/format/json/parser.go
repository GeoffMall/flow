package json

import (
	"encoding/json"
	"errors"
	"io"
)

// Parser implements format.Parser for JSON format.
// It supports:
//   - Streaming array elements individually
//   - Concatenated JSON documents
//   - Standard JSON objects and primitives
type Parser struct {
	dec *json.Decoder
}

// NewParser creates a new JSON streaming parser.
func NewParser(r io.Reader) *Parser {
	dec := json.NewDecoder(r)
	dec.UseNumber() // Preserve number precision
	return &Parser{dec: dec}
}

// ForEach streams JSON values and calls fn for each document.
// If the first value is an array, each array element is streamed separately.
// Supports concatenated JSON documents.
func (p *Parser) ForEach(fn func(any) error) error {
	// Read first document
	rm, err := p.decodeFirstMessage()
	if err != nil {
		return err
	}

	// If first document is an array, stream its elements
	if p.isArray(rm) {
		if err := p.processArrayElements(rm, fn); err != nil {
			return err
		}
		// Continue with any concatenated documents
		return p.processConcatenatedDocuments(fn)
	}

	// Process single value
	if err := p.processRawMessage(rm, fn); err != nil {
		return err
	}

	// Continue with any concatenated documents
	return p.processConcatenatedDocuments(fn)
}

// decodeFirstMessage reads the first JSON message from the decoder
func (p *Parser) decodeFirstMessage() (json.RawMessage, error) {
	var rm json.RawMessage
	if err := p.dec.Decode(&rm); err != nil {
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

// processArrayElements streams individual elements from a JSON array
func (p *Parser) processArrayElements(rm json.RawMessage, fn func(any) error) error {
	var arr []any
	if err := json.Unmarshal(rm, &arr); err != nil {
		return err
	}

	for i := range arr {
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

	return fn(v)
}

// processConcatenatedDocuments continues reading concatenated JSON documents
func (p *Parser) processConcatenatedDocuments(fn func(any) error) error {
	for {
		var rm json.RawMessage
		if err := p.dec.Decode(&rm); err != nil {
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
