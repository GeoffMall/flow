package json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/GeoffMall/flow/internal/format"
	"github.com/stretchr/testify/assert"
)

func TestParser_SingleObject(t *testing.T) {
	input := `{"name": "Alice", "age": 30}`
	parser := NewParser(strings.NewReader(input))

	var docs []any
	err := parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, docs, 1)

	obj, ok := docs[0].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Alice", obj["name"])
	assert.Equal(t, float64(30), obj["age"])
}

func TestParser_Array(t *testing.T) {
	input := `[{"id": 1}, {"id": 2}, {"id": 3}]`
	parser := NewParser(strings.NewReader(input))

	var docs []any
	err := parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, docs, 3, "array elements should be streamed individually")

	for i, doc := range docs {
		obj, ok := doc.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, float64(i+1), obj["id"])
	}
}

func TestParser_ConcatenatedDocuments(t *testing.T) {
	input := `{"id": 1}
{"id": 2}
{"id": 3}`
	parser := NewParser(strings.NewReader(input))

	var docs []any
	err := parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, docs, 3)

	for i, doc := range docs {
		obj, ok := doc.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, float64(i+1), obj["id"])
	}
}

func TestFormatter_Compact(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{
		Compact: true,
		Color:   false,
	})

	doc := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"name":"Alice"`)
	assert.Contains(t, output, `"age":30`)
	assert.NotContains(t, output, "  ") // No indentation in compact mode
}

func TestFormatter_Pretty(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{
		Compact: false,
		Color:   false,
	})

	doc := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"name": "Alice"`)
	assert.Contains(t, output, `"age": 30`)
	assert.Contains(t, output, "  ") // Indentation in pretty mode
}

func TestFormatter_Color(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{
		Compact: false,
		Color:   true,
	})

	doc := map[string]any{
		"name":    "Alice",
		"isAdmin": true,
		"age":     30,
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "\x1b[") // Should contain ANSI escape codes
	assert.Contains(t, output, colKey)  // Key color
	assert.Contains(t, output, colStr)  // String color
	assert.Contains(t, output, colNum)  // Number color
}

func TestFormatter_ArraysWithColor(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{
		Compact: false,
		Color:   true,
	})

	doc := map[string]any{
		"items": []any{1, 2, 3},
		"tags":  []any{"go", "json"},
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "[")     // Array start
	assert.Contains(t, output, "]")     // Array end
	assert.Contains(t, output, "\x1b[") // ANSI codes
}

func TestFormatter_NestedStructures(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{
		Compact: false,
		Color:   true,
	})

	doc := map[string]any{
		"user": map[string]any{
			"name":  "Alice",
			"roles": []any{"admin", "user"},
		},
		"active": true,
		"count":  42,
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "admin")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "true")
}

func TestFormatter_NullValue(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{
		Compact: false,
		Color:   true,
	})

	doc := map[string]any{
		"value": nil,
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "null")
}

func TestParser_InvalidJSON(t *testing.T) {
	input := `{invalid json}`
	parser := NewParser(strings.NewReader(input))

	err := parser.ForEach(func(doc any) error {
		return nil
	})

	assert.Error(t, err, "should fail on invalid JSON")
}

func TestParser_EmptyArray(t *testing.T) {
	input := `[]`
	parser := NewParser(strings.NewReader(input))

	var docs []any
	err := parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, docs, 0, "empty array should produce no documents")
}

func TestParser_MixedTypes(t *testing.T) {
	input := `[1, "string", true, null, {"key": "value"}]`
	parser := NewParser(strings.NewReader(input))

	var docs []any
	err := parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, docs, 5)

	assert.Equal(t, float64(1), docs[0])
	assert.Equal(t, "string", docs[1])
	assert.Equal(t, true, docs[2])
	assert.Nil(t, docs[3])
	assert.IsType(t, map[string]any{}, docs[4])
}

func TestFormat_Integration(t *testing.T) {
	fmt := &Format{}

	// Test name
	assert.Equal(t, "json", fmt.Name())

	// Test parser
	input := strings.NewReader(`{"test": true}`)
	parser, err := fmt.NewParser(input)
	assert.NoError(t, err)

	var docs []any
	err = parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, docs, 1)

	// Test formatter
	buf := &bytes.Buffer{}
	formatter := fmt.NewFormatter(buf, format.FormatterOptions{})
	err = formatter.Write(docs[0])
	assert.NoError(t, err)
	err = formatter.Close()
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `"test"`)
}
