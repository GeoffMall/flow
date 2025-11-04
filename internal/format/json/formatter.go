package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/GeoffMall/flow/internal/format"
)

// Formatter implements format.Formatter for JSON output.
type Formatter struct {
	w       io.Writer
	color   bool
	compact bool
}

// NewFormatter creates a new JSON formatter with the given options.
func NewFormatter(w io.Writer, opts format.FormatterOptions) *Formatter {
	return &Formatter{
		w:       w,
		color:   opts.Color,
		compact: opts.Compact,
	}
}

// Write outputs a single JSON document with formatting.
// Respects compact and color options.
func (f *Formatter) Write(doc any) error {
	var b []byte
	var err error

	if f.compact {
		// Compact JSON - single line
		b, err = json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
	} else {
		// Pretty print with 2-space indentation
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		if err := enc.Encode(doc); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		b = buf.Bytes()
		// Encoder adds a trailing newline - keep it for readability
	}

	// Apply colorization if requested
	if f.color {
		b = colorizeJSON(b)
	}

	// Write to output
	_, err = f.w.Write(b)
	if err != nil {
		return err
	}

	// Ensure newline for compact mode (pretty already has one)
	if f.compact && (len(b) == 0 || b[len(b)-1] != '\n') {
		if _, err := f.w.Write([]byte{'\n'}); err != nil {
			return err
		}
	}

	return nil
}

// Close flushes any buffered data (no-op for JSON)
func (f *Formatter) Close() error {
	return nil
}

// colorizeJSON adds ANSI color codes to JSON bytes for terminal display.
// Uses a state machine to identify and color:
//   - Object keys (blue)
//   - String values (green)
//   - Numbers (orange)
//   - Booleans and null (purple)
//   - Punctuation (gray)
func colorizeJSON(in []byte) []byte {
	colorizer := newJSONColorizer(in)
	return colorizer.colorize()
}

// ANSI color codes
const (
	colReset   = "\x1b[0m"
	colKey     = "\x1b[38;5;33m"  // blue for keys
	colStr     = "\x1b[38;5;34m"  // green for strings
	colNum     = "\x1b[38;5;214m" // orange for numbers
	colBoolNil = "\x1b[38;5;135m" // purple for true/false/null
	colPunct   = "\x1b[38;5;240m" // gray for punctuation
)

type jsonColorizer struct {
	input  []byte
	output []byte
	stack  []objState
	inStr  bool
	esc    bool
}

type objState struct {
	expectKey bool
}

func newJSONColorizer(in []byte) *jsonColorizer {
	return &jsonColorizer{
		input:  in,
		output: make([]byte, 0, len(in)+len(in)/4), // Extra space for color codes
		stack:  make([]objState, 0),
	}
}

func (c *jsonColorizer) colorize() []byte {
	for i := 0; i < len(c.input); i++ {
		b := c.input[i]

		if c.inStr {
			c.processStringChar(b)
			continue
		}

		c.processNonStringChar(b, &i)
	}

	c.ensureNewline()
	return c.output
}

func (c *jsonColorizer) processStringChar(b byte) {
	c.writeByte(b)

	if c.esc {
		c.esc = false
		return
	}

	if b == '\\' {
		c.esc = true
		return
	}

	if b == '"' {
		c.write(colReset)
		c.inStr = false
	}
}

func (c *jsonColorizer) processNonStringChar(b byte, i *int) {
	switch b {
	case '{':
		c.handleObjectStart()
	case '}':
		c.handleObjectEnd()
	case '[':
		c.handleArrayStart()
	case ']':
		c.handleArrayEnd()
	case ':', ',':
		c.handlePunctuation(b)
	case '"':
		c.handleStringStart()
	case 't', 'f', 'n':
		c.handleBooleanOrNull(b, i)
	default:
		c.handleDefault(b, i)
	}
}

func (c *jsonColorizer) handleBooleanOrNull(b byte, i *int) {
	switch b {
	case 't':
		c.handleKeyword(i, "true")
	case 'f':
		c.handleKeyword(i, "false")
	case 'n':
		c.handleKeyword(i, "null")
	}
}

func (c *jsonColorizer) handleObjectStart() {
	c.writePunctuation('{')
	c.pushState(objState{expectKey: true})
}

func (c *jsonColorizer) handleObjectEnd() {
	c.writePunctuation('}')
	c.popState()
}

func (c *jsonColorizer) handleArrayStart() {
	c.writePunctuation('[')
	c.pushState(objState{expectKey: false})
}

func (c *jsonColorizer) handleArrayEnd() {
	c.writePunctuation(']')
	c.popState()
}

func (c *jsonColorizer) handlePunctuation(b byte) {
	c.writePunctuation(b)

	if b == ':' {
		c.setExpectKey(false)
	} else if b == ',' {
		c.setExpectKey(true)
	}
}

func (c *jsonColorizer) handleStringStart() {
	if c.isExpectingKey() {
		c.write(colKey)
	} else {
		c.write(colStr)
	}

	c.writeByte('"')
	c.inStr = true
}

func (c *jsonColorizer) handleKeyword(i *int, keyword string) {
	if tryWord(c.input, i, keyword, &c.output, colBoolNil) {
		return
	}
	c.writeByte(c.input[*i])
}

func (c *jsonColorizer) handleDefault(b byte, i *int) {
	if isDigitOrNumberChar(b) {
		c.handleNumber(i)
	} else {
		c.writeByte(b)
	}
}

func (c *jsonColorizer) handleNumber(i *int) {
	c.write(colNum)

	j := *i
	for j < len(c.input) && isDigitOrNumberChar(c.input[j]) {
		j++
	}

	c.output = append(c.output, c.input[*i:j]...)
	c.write(colReset)
	*i = j - 1
}

// Helper methods
func (c *jsonColorizer) write(s string) {
	c.output = append(c.output, s...)
}

func (c *jsonColorizer) writeByte(b byte) {
	c.output = append(c.output, b)
}

func (c *jsonColorizer) writePunctuation(b byte) {
	c.write(colPunct)
	c.writeByte(b)
	c.write(colReset)
}

func (c *jsonColorizer) pushState(s objState) {
	c.stack = append(c.stack, s)
}

func (c *jsonColorizer) popState() {
	if len(c.stack) > 0 {
		c.stack = c.stack[:len(c.stack)-1]
	}
}

func (c *jsonColorizer) topState() *objState {
	if len(c.stack) == 0 {
		return nil
	}
	return &c.stack[len(c.stack)-1]
}

func (c *jsonColorizer) isExpectingKey() bool {
	if st := c.topState(); st != nil {
		return st.expectKey
	}
	return false
}

func (c *jsonColorizer) setExpectKey(expectKey bool) {
	if st := c.topState(); st != nil {
		st.expectKey = expectKey
	}
}

func (c *jsonColorizer) ensureNewline() {
	if len(c.output) == 0 || c.output[len(c.output)-1] != '\n' {
		c.output = append(c.output, '\n')
	}
}

func isDigitOrNumberChar(b byte) bool {
	switch b {
	case '-', '+', '.', 'e', 'E':
		return true
	default:
		return b >= '0' && b <= '9'
	}
}

func tryWord(in []byte, i *int, word string, out *[]byte, color string) bool {
	if hasWordAt(in, *i, word) {
		*out = append(*out, color...)
		*out = append(*out, in[*i:*i+len(word)]...)
		*out = append(*out, colReset...)
		*i += len(word) - 1
		return true
	}
	return false
}

func hasWordAt(b []byte, i int, w string) bool {
	if i+len(w) > len(b) {
		return false
	}

	for j := 0; j < len(w); j++ {
		if b[i+j] != w[j] {
			return false
		}
	}

	return true
}
