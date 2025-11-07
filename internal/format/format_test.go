package format

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing

type mockParser struct {
	docs []any
	err  error
}

func (m *mockParser) ForEach(fn func(doc any) error) error {
	if m.err != nil {
		return m.err
	}
	for _, doc := range m.docs {
		if err := fn(doc); err != nil {
			return err
		}
	}
	return nil
}

type mockFormatter struct {
	buf    *bytes.Buffer
	closed bool
}

func (m *mockFormatter) Write(doc any) error {
	_, err := m.buf.WriteString("doc\n")
	return err
}

func (m *mockFormatter) Close() error {
	m.closed = true
	return nil
}

type mockFormat struct {
	name      string
	parser    Parser
	formatter Formatter
}

func (m *mockFormat) Name() string {
	return m.name
}

func (m *mockFormat) NewParser(r io.Reader) (Parser, error) {
	return m.parser, nil
}

func (m *mockFormat) NewFormatter(w io.Writer, opts FormatterOptions) Formatter {
	return m.formatter
}

// Tests

func TestRegister(t *testing.T) {
	// Clear registry for testing
	registryMu.Lock()
	registry = make(map[string]Format)
	registryMu.Unlock()

	mock := &mockFormat{name: "test"}
	Register(mock)

	got, err := Get("test")
	assert.NoError(t, err)
	assert.Equal(t, mock, got)
}

func TestGet_NotFound(t *testing.T) {
	// Clear registry for testing
	registryMu.Lock()
	registry = make(map[string]Format)
	registryMu.Unlock()

	_, err := Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format")
}

func TestList(t *testing.T) {
	// Clear registry for testing
	registryMu.Lock()
	registry = make(map[string]Format)
	registryMu.Unlock()

	Register(&mockFormat{name: "format1"})
	Register(&mockFormat{name: "format2"})

	names := List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "format1")
	assert.Contains(t, names, "format2")
}

func TestFormatterClose(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{buf: buf}

	err := formatter.Write(map[string]any{"key": "value"})
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)
	assert.True(t, formatter.closed)
}
