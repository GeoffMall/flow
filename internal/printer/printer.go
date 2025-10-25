package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/GeoffMall/flow/internal/parser"
)

// Options controls printer behavior.
type Options struct {
	ToFormat string // "json" | "yaml" | "" (defaults to json)
	Color    bool   // colorize JSON output with ANSI
	Compact  bool   // minified JSON (ignored for YAML)
	Writer   io.Writer
}

// Printer emits one output document per Write() call.
type Printer struct {
	w      io.Writer
	format string // "json" | "yaml"
	opt    Options
}

// New returns a new Printer with sane defaults.
// If opts.Writer is nil, it uses os.Stdout.
// If opts.ToFormat is empty or invalid, it falls back to "json".
func New(opts Options) *Printer {
	w := opts.Writer
	if w == nil {
		w = os.Stdout
	}
	format := "json"
	if opts.ToFormat == "yaml" {
		format = "yaml"
	}
	return &Printer{
		w:      w,
		format: format,
		opt:    opts,
	}
}

// Write prints a single structured value as one document.
// - JSON: respects Compact/Color
// - YAML: pretty prints with 2-space indent (no color)
func (p *Printer) Write(v any) error {
	switch p.format {
	case "yaml":
		return p.writeYAML(v)
	default: // json
		return p.writeJSON(v)
	}
}

func (p *Printer) writeJSON(v any) error {
	var b []byte
	var err error

	if p.opt.Compact {
		b, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
	} else {
		// Pretty print
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		b = buf.Bytes()
		// Encoder adds a trailing newline; keep it (ensures one doc per line).
	}

	if p.opt.Color {
		cb := colorizeJSON(b)
		_, err = p.w.Write(cb)
		return err
	}

	_, err = p.w.Write(b)
	if err != nil {
		return err
	}

	// Ensure new line for compact mode (pretty already has one).
	if p.opt.Compact {
		if len(b) == 0 || b[len(b)-1] != '\n' {
			_, _ = p.w.Write([]byte{'\n'})
		}
	}

	return nil
}

func (p *Printer) writeYAML(v any) error {
	enc := parser.NewYAMLEncoder(p.w)
	defer enc.Close()
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("yaml encode: %w", err)
	}
	return nil
}
