// Package yaml implements YAML format support for flow.
// It provides parsing and formatting of YAML data with:
//   - Streaming document processing (--- separated documents)
//   - Normalization to JSON-compatible types
//   - Pretty-printed output with 2-space indentation
package yaml

import (
	"io"

	"github.com/GeoffMall/flow/internal/format"
)

// Format implements format.Format for YAML.
type Format struct{}

// Name returns the format identifier.
func (f *Format) Name() string {
	return "yaml"
}

// NewParser creates a new YAML streaming parser.
func (f *Format) NewParser(r io.Reader) (format.Parser, error) {
	return NewParser(r), nil
}

// NewFormatter creates a new YAML formatter.
func (f *Format) NewFormatter(w io.Writer, opts format.FormatterOptions) format.Formatter {
	return NewFormatter(w, opts)
}

// Register the YAML format on package initialization
//
//nolint:gochecknoinits // Required for automatic format registration
func init() {
	format.Register(&Format{})
}
