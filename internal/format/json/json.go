// Package json implements JSON format support for flow.
// It provides detection, parsing, and formatting of JSON data with:
//   - Streaming array element processing
//   - Concatenated JSON document support
//   - ANSI color output for terminal display
//   - Compact and pretty-print modes
package json

import (
	"io"

	"github.com/GeoffMall/flow/internal/format"
)

// Format implements format.Format for JSON.
type Format struct{}

// Name returns the format identifier.
func (f *Format) Name() string {
	return "json"
}

// Detector returns a JSON format detector.
func (f *Format) Detector() format.Detector {
	return &Detector{}
}

// NewParser creates a new JSON streaming parser.
func (f *Format) NewParser(r io.Reader) (format.Parser, error) {
	return NewParser(r), nil
}

// NewFormatter creates a new JSON formatter.
func (f *Format) NewFormatter(w io.Writer, opts format.FormatterOptions) format.Formatter {
	return NewFormatter(w, opts)
}

// Register the JSON format on package initialization
//
//nolint:gochecknoinits // Required for automatic format registration
func init() {
	format.Register(&Format{})
}
