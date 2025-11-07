package avro

import (
	"fmt"
	"io"

	"github.com/hamba/avro/v2/ocf"
)

// Parser implements the format.Parser interface for Avro OCF (Object Container Files).
// It streams records from an Avro file without buffering the entire file into memory.
type Parser struct {
	decoder *ocf.Decoder
}

// NewParser creates a new Avro parser that reads from the given reader.
// The reader must contain a valid Avro OCF file with embedded schema.
func NewParser(r io.Reader) (*Parser, error) {
	dec, err := ocf.NewDecoder(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create avro decoder: %w", err)
	}

	return &Parser{
		decoder: dec,
	}, nil
}

// ForEach iterates over all records in the Avro file, calling fn for each record.
// Records are decoded into map[string]any for format-agnostic processing.
// Iteration stops when:
// - All records have been processed (returns nil)
// - The callback fn returns an error (returns that error)
// - The decoder encounters an error (returns that error)
func (p *Parser) ForEach(fn func(doc any) error) error {
	for p.decoder.HasNext() {
		// Decode into a generic map for format-agnostic operations
		var record map[string]any
		if err := p.decoder.Decode(&record); err != nil {
			return fmt.Errorf("failed to decode avro record: %w", err)
		}

		// Call the callback with the decoded record
		if err := fn(record); err != nil {
			return err
		}
	}

	// Check for decoder errors after iteration completes
	if err := p.decoder.Error(); err != nil {
		return fmt.Errorf("avro decoder error: %w", err)
	}

	return nil
}
