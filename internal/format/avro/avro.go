package avro

import (
	"fmt"
	"io"

	"github.com/GeoffMall/flow/internal/format"
)

// Format implements the format.Format interface for Apache Avro.
// This implementation provides read-only support for Avro OCF (Object Container Files).
type Format struct{}

// Name returns the format identifier used in CLI flags (-from avro).
func (f *Format) Name() string {
	return "avro"
}

// NewParser creates a new parser for reading Avro OCF files.
func (f *Format) NewParser(r io.Reader) (format.Parser, error) {
	return NewParser(r)
}

// NewFormatter creates a formatter for writing Avro files.
// Currently not implemented as this feature only supports reading Avro files.
func (f *Format) NewFormatter(w io.Writer, opts format.FormatterOptions) format.Formatter {
	// Avro writing not supported in this implementation
	// Return nil - the caller should check for this and provide a clear error
	panic(fmt.Sprintf("avro format does not support writing (formatter not implemented)"))
}

//nolint:gochecknoinits // Init required for format registration
func init() {
	// Register the Avro format with the global format registry
	// This enables automatic discovery via -from avro flag
	format.Register(&Format{})
}
