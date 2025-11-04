// Package format provides abstractions for different data formats (JSON, YAML, CSV, etc.).
//
// # Architecture
//
// The format package defines core interfaces that all formats must implement:
//
//  1. Detector - Analyzes input bytes to determine if data matches a format
//  2. Parser - Streams documents from input without full buffering
//  3. Formatter - Writes documents to output with optional styling
//  4. Format - Combines the above into a complete format implementation
//
// # Adding a New Format
//
// To add support for a new format:
//
//  1. Create a new package under internal/format/yourformat/
//  2. Implement Detector, Parser, and Formatter interfaces
//  3. Create a Format implementation that wires them together
//  4. Register the format in init() using format.Register()
//
// See internal/format/json/ for a complete reference implementation.
//
// # Streaming Semantics
//
// Parsers must stream documents without loading entire input into memory.
// For array-based formats (JSON arrays), each element should be streamed.
// For document-based formats (YAML with ---), each document is streamed.
// For row-based formats (CSV), each row should be streamed as a document.
package format

import (
	"io"
)

// Format represents a data format (JSON, YAML, CSV, Parquet, Avro, etc.).
// It provides format detection, parsing, and formatting capabilities.
type Format interface {
	// Name returns the format identifier (e.g., "json", "yaml", "csv")
	Name() string

	// Detector returns a detector that can identify this format from input bytes
	Detector() Detector

	// NewParser creates a streaming parser for this format
	NewParser(r io.Reader) (Parser, error)

	// NewFormatter creates a formatter for output with the given options
	NewFormatter(w io.Writer, opts FormatterOptions) Formatter
}

// Detector analyzes input bytes to determine if they match a specific format.
// Detectors are used during auto-detection to identify the input format.
type Detector interface {
	// Detect analyzes the given bytes and returns a confidence score (0-100).
	// Higher scores indicate stronger confidence that the data matches this format.
	// Returns 0 if the data definitely does not match this format.
	// Returns an error only if detection itself fails (not if format doesn't match).
	Detect(peek []byte) (confidence int, err error)
}

// Parser streams documents from input data without loading everything into memory.
// Parsers normalize data into Go's standard types (map[string]any, []any, primitives).
type Parser interface {
	// ForEach calls fn for each document/row in the input stream.
	// For JSON arrays, each array element is treated as a separate document.
	// For YAML, each document separated by --- is processed individually.
	// For CSV, each row (after headers) is converted to map[string]any.
	// Processing stops when fn returns an error or end of input is reached.
	ForEach(fn func(doc any) error) error
}

// Formatter writes documents to output, with optional formatting and styling.
type Formatter interface {
	// Write outputs a single document/row.
	// The document should be in normalized form (map[string]any, []any, etc.).
	Write(doc any) error

	// Close flushes any buffered data and releases resources.
	// Must be called when done writing.
	Close() error
}

// FormatterOptions holds common formatting options applicable across formats.
type FormatterOptions struct {
	// Color enables ANSI color codes in output (for terminal display)
	Color bool

	// Compact removes unnecessary whitespace for minimal output size
	Compact bool
}
