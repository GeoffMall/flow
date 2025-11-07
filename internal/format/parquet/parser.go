package parquet

import (
	"fmt"
	"io"
	"os"

	"github.com/parquet-go/parquet-go"
)

// Parser implements the format.Parser interface for Apache Parquet files.
// It streams rows from a Parquet file without buffering the entire file into memory.
//
// Note: Parquet requires seekable input (actual files), so io.Reader is not sufficient.
// The NewParser function will attempt to get the underlying *os.File if possible.
type Parser struct {
	file   *parquet.File
	reader *parquet.Reader
}

// NewParser creates a new Parquet parser that reads from the given reader.
// The reader must be an *os.File or provide seekable access to the Parquet file.
// Returns an error if the reader is not seekable (e.g., stdin).
func NewParser(r io.Reader) (*Parser, error) {
	// Parquet requires seekable input - try to get underlying file
	osFile, ok := r.(*os.File)
	if !ok {
		return nil, fmt.Errorf("parquet format requires seekable file input (not stdin or pipe)")
	}

	// Get file info for size
	stat, err := osFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat parquet file: %w", err)
	}

	// Open parquet file
	pf, err := parquet.OpenFile(osFile, stat.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}

	// Create a generic reader
	reader := parquet.NewReader(pf)

	return &Parser{
		file:   pf,
		reader: reader,
	}, nil
}

// ForEach iterates over all rows in the Parquet file, calling fn for each row.
// Rows are returned as map[string]any for format-agnostic processing.
// Iteration stops when:
// - All rows have been processed (returns nil)
// - The callback fn returns an error (returns that error)
// - The reader encounters an error (returns that error)
func (p *Parser) ForEach(fn func(doc any) error) error {
	// Read rows one at a time
	for {
		// Read next row as a generic map
		row := make(map[string]any)
		err := p.reader.Read(&row)
		if err != nil {
			if err == io.EOF {
				// End of file reached
				break
			}
			return fmt.Errorf("failed to read parquet row: %w", err)
		}

		// Call the callback with the row
		if err := fn(row); err != nil {
			return err
		}
	}

	return nil
}
