package yaml

import (
	"fmt"
	"io"

	"github.com/GeoffMall/flow/internal/format"
	"gopkg.in/yaml.v3"
)

// Formatter implements format.Formatter for YAML output.
type Formatter struct {
	enc *yaml.Encoder
}

// NewFormatter creates a new YAML formatter.
// Note: Color option is ignored for YAML (YAML doesn't typically use color codes).
func NewFormatter(w io.Writer, opts format.FormatterOptions) *Formatter {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2) // Standard 2-space YAML indentation
	return &Formatter{
		enc: enc,
	}
}

// Write outputs a single YAML document.
// Each call writes a document with trailing newline.
func (f *Formatter) Write(doc any) error {
	if err := f.enc.Encode(doc); err != nil {
		return fmt.Errorf("yaml encode: %w", err)
	}
	return nil
}

// Close flushes the encoder and releases resources.
// Must be called when done writing.
func (f *Formatter) Close() error {
	if f.enc != nil {
		return f.enc.Close()
	}
	return nil
}
