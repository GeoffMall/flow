package parquet

import (
	"fmt"
	"io"

	"github.com/GeoffMall/flow/internal/format"
)

// Format implements the format.Format interface for Apache Parquet.
// This implementation provides read-only support for Parquet files.
type Format struct{}

// Name returns the format identifier used in CLI flags (-from parquet).
func (f *Format) Name() string {
	return "parquet"
}

// NewParser creates a new parser for reading Parquet files.
// Note: Parquet requires seekable file input. Passing stdin will result in an error.
func (f *Format) NewParser(r io.Reader) (format.Parser, error) {
	return NewParser(r)
}

// NewFormatter creates a formatter for writing Parquet files.
// Currently not implemented as this feature only supports reading Parquet files.
func (f *Format) NewFormatter(w io.Writer, opts format.FormatterOptions) format.Formatter {
	// Parquet writing not supported in this implementation
	// Return nil - the caller should check for this and provide a clear error
	panic(fmt.Sprintf("parquet format does not support writing (formatter not implemented)"))
}

//nolint:gochecknoinits // Init required for format registration
func init() {
	// Register the Parquet format with the global format registry
	// This enables automatic discovery via -from parquet flag
	format.Register(&Format{})
}
