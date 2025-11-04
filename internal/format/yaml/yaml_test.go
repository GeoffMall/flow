package yaml

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
		wantIsYAML bool
	}{
		{
			name:       "yaml directive",
			input:      "%YAML 1.2\n---\nkey: value",
			wantScore:  100,
			wantIsYAML: true,
		},
		{
			name:       "document separator",
			input:      "---\nkey: value",
			wantScore:  100,
			wantIsYAML: true,
		},
		{
			name:       "key value style",
			input:      "name: Alice\nage: 30",
			wantScore:  90,
			wantIsYAML: true,
		},
		{
			name:       "whitespace then yaml",
			input:      "  \n  key: value",
			wantScore:  90,
			wantIsYAML: true,
		},
		{
			name:       "json object",
			input:      `{"key": "value"}`,
			wantScore:  0,
			wantIsYAML: false,
		},
		{
			name:       "json array",
			input:      `[1, 2, 3]`,
			wantScore:  0,
			wantIsYAML: false,
		},
		{
			name:       "empty input",
			input:      "",
			wantScore:  0,
			wantIsYAML: false,
		},
	}

	detector := &Detector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := detector.Detect([]byte(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.wantScore, score)

			if tt.wantIsYAML {
				assert.Greater(t, score, 0, "should detect as YAML")
			} else {
				assert.Equal(t, 0, score, "should not detect as YAML")
			}
		})
	}
}

func TestParser_SingleDocument(t *testing.T) {
	input := `name: Alice
age: 30`
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
	assert.Equal(t, 30, obj["age"])
}

func TestParser_MultipleDocuments(t *testing.T) {
	input := `---
name: Alice
---
name: Bob
---
name: Charlie`
	parser := NewParser(strings.NewReader(input))

	var docs []any
	err := parser.ForEach(func(doc any) error {
		docs = append(docs, doc)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, docs, 3)

	names := []string{"Alice", "Bob", "Charlie"}
	for i, doc := range docs {
		obj, ok := doc.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, names[i], obj["name"])
	}
}

func TestParser_ArrayInYAML(t *testing.T) {
	input := `items:
  - id: 1
    name: first
  - id: 2
    name: second`
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

	items, ok := obj["items"].([]any)
	assert.True(t, ok)
	assert.Len(t, items, 2)
}

func TestParser_Normalization(t *testing.T) {
	// YAML allows non-string keys; we normalize to strings
	input := `123: numeric key
true: boolean key`
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
	assert.Equal(t, "numeric key", obj["123"])
	assert.Equal(t, "boolean key", obj["true"])
}

func TestFormatter(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, format.FormatterOptions{})

	doc := map[string]any{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}

	err := formatter.Write(doc)
	assert.NoError(t, err)

	err = formatter.Close()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name: Alice")
	assert.Contains(t, output, "age: 30")
	assert.Contains(t, output, "active: true")
}

func TestFormat_Integration(t *testing.T) {
	fmt := &Format{}

	// Test name
	assert.Equal(t, "yaml", fmt.Name())

	// Test detector
	detector := fmt.Detector()
	score, err := detector.Detect([]byte("key: value"))
	assert.NoError(t, err)
	assert.Greater(t, score, 0)

	// Test parser
	input := strings.NewReader("name: Alice\nage: 30")
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

	assert.Contains(t, buf.String(), "name:")
	assert.Contains(t, buf.String(), "Alice")
}
