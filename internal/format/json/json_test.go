package json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/GeoffMall/flow/internal/format"
	"github.com/stretchr/testify/assert"
)

//nolint:funlen // Test table is intentionally comprehensive
func TestDetector(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantScore  int
		wantIsJSON bool
	}{
		{
			name:       "object start",
			input:      `{"key": "value"}`,
			wantScore:  100,
			wantIsJSON: true,
		},
		{
			name:       "array start",
			input:      `[1, 2, 3]`,
			wantScore:  100,
			wantIsJSON: true,
		},
		{
			name:       "empty input",
			input:      "",
			wantScore:  80,
			wantIsJSON: true,
		},
		{
			name:       "whitespace then object",
			input:      "  \n  {\"key\": \"value\"}",
			wantScore:  100,
			wantIsJSON: true,
		},
		{
			name:       "yaml directive",
			input:      "%YAML 1.2\n---",
			wantScore:  0,
			wantIsJSON: false,
		},
		{
			name:       "yaml document separator",
			input:      "---\nkey: value",
			wantScore:  0,
			wantIsJSON: false,
		},
		{
			name:       "yaml style key value",
			input:      "key: value\nother: thing",
			wantScore:  0,
			wantIsJSON: false,
		},
		{
			name:       "negative number",
			input:      "-123",
			wantScore:  50,
			wantIsJSON: true,
		},
		{
			name:       "string value",
			input:      `"hello"`,
			wantScore:  50,
			wantIsJSON: true,
		},
	}

	detector := &Detector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := detector.Detect([]byte(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.wantScore, score)

			if tt.wantIsJSON {
				assert.Greater(t, score, 0, "should detect as possible JSON")
			} else {
				assert.Equal(t, 0, score, "should not detect as JSON")
			}
		})
	}
}

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

func TestFormat_Integration(t *testing.T) {
	fmt := &Format{}

	// Test name
	assert.Equal(t, "json", fmt.Name())

	// Test detector
	detector := fmt.Detector()
	score, err := detector.Detect([]byte(`{"test": true}`))
	assert.NoError(t, err)
	assert.Equal(t, 100, score)

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
