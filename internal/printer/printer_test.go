package printer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Test helper functions for color assertions

// assertHasColor checks that the output contains ANSI color codes
func assertHasColor(t *testing.T, output string) {
	t.Helper()
	assert.Contains(t, output, "\x1b[", "output should contain ANSI escape codes")
	assert.Contains(t, output, colReset, "output should contain color reset codes")
}

// assertNoColor checks that the output does not contain ANSI color codes
func assertNoColor(t *testing.T, output string) {
	t.Helper()
	assert.NotContains(t, output, "\x1b[", "output should not contain ANSI escape codes")
}

// assertHasColorType checks that a specific color type is present in the output
func assertHasColorType(t *testing.T, output, colorCode, description string) {
	t.Helper()
	assert.Contains(t, output, colorCode, "output should contain %s", description)
}

func TestNew_NonNil(t *testing.T) {
	p := New(Options{})
	assert.NotNil(t, p)
}

func TestNew_DefaultWriter(t *testing.T) {
	p := New(Options{})
	assert.NotNil(t, p.w)
}

func TestNew_CustomWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})
	assert.Equal(t, buf, p.w)
}

func TestNew_DefaultFormat(t *testing.T) {
	p := New(Options{})
	assert.Equal(t, "json", p.format)
}

func TestNew_YAMLFormat(t *testing.T) {
	p := New(Options{ToFormat: "yaml"})
	assert.Equal(t, "yaml", p.format)
}

func TestNew_InvalidFormatDefaultsToJSON(t *testing.T) {
	p := New(Options{ToFormat: "invalid"})
	assert.Equal(t, "json", p.format)
}

// JSON output tests
func TestWrite_JSON_SimpleObject(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	input := map[string]any{"name": "alice", "age": 30}
	err := p.Write(input)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "age")
	assert.Contains(t, output, "30")

	// Verify it's valid JSON
	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
}

func TestWrite_JSON_Array(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	input := []any{1, 2, 3, "four", true}
	err := p.Write(input)
	require.NoError(t, err)

	var result []any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 5)
	assert.Equal(t, float64(1), result[0])
	assert.Equal(t, "four", result[3])
	assert.Equal(t, true, result[4])
}

func TestWrite_JSON_NestedStructure(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	input := map[string]any{
		"user": map[string]any{
			"name":  "alice",
			"email": "alice@example.com",
		},
		"tags": []any{"golang", "cli"},
	}
	err := p.Write(input)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	user := result["user"].(map[string]any)
	assert.Equal(t, "alice", user["name"])

	tags := result["tags"].([]any)
	assert.Len(t, tags, 2)
}

func TestWrite_JSON_PrettyPrint(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, Compact: false})

	input := map[string]any{"name": "alice", "age": 30}
	err := p.Write(input)
	require.NoError(t, err)

	output := buf.String()
	// Pretty print should have indentation and newlines
	assert.Contains(t, output, "\n")
	assert.Contains(t, output, "  ")
}

func TestWrite_JSON_Compact(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, Compact: true})

	input := map[string]any{"name": "alice", "age": 30}
	err := p.Write(input)
	require.NoError(t, err)

	output := strings.TrimSpace(buf.String())
	// Compact should be single line (no internal newlines)
	assert.NotContains(t, strings.TrimRight(output, "\n"), "\n")
}

func TestWrite_JSON_CompactEndsWithNewline(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, Compact: true})

	input := map[string]any{"name": "alice"}
	err := p.Write(input)
	require.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.HasSuffix(output, "\n"))
}

func TestWrite_JSON_MultipleDocuments(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, Compact: true})

	err := p.Write(map[string]any{"first": 1})
	require.NoError(t, err)
	err = p.Write(map[string]any{"second": 2})
	require.NoError(t, err)
	err = p.Write(map[string]any{"third": 3})
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 3)
}

func TestWrite_JSON_AllTypes(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	input := map[string]any{
		"string":  "hello",
		"number":  42.5,
		"integer": 10,
		"boolean": true,
		"null":    nil,
		"array":   []any{1, 2, 3},
		"object":  map[string]any{"nested": "value"},
	}
	err := p.Write(input)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result["string"])
	assert.Equal(t, 42.5, result["number"])
	assert.Equal(t, true, result["boolean"])
	assert.Nil(t, result["null"])
}

func TestWrite_JSON_EmptyObject(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	err := p.Write(map[string]any{})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestWrite_JSON_EmptyArray(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	err := p.Write([]any{})
	require.NoError(t, err)

	var result []any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestWrite_JSON_ScalarString(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	err := p.Write("hello")
	require.NoError(t, err)

	var result string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestWrite_JSON_ScalarNumber(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	err := p.Write(42)
	require.NoError(t, err)

	var result int
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestWrite_JSON_ScalarBoolean(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	err := p.Write(true)
	require.NoError(t, err)

	var result bool
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestWrite_JSON_ScalarNull(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf})

	err := p.Write(nil)
	require.NoError(t, err)

	output := strings.TrimSpace(buf.String())
	assert.Equal(t, "null", output)
}

// Color tests
func TestWrite_JSON_WithColor(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, Color: true})

	input := map[string]any{"name": "alice", "age": 30}
	err := p.Write(input)
	require.NoError(t, err)

	assertHasColor(t, buf.String())
}

func TestWrite_JSON_WithoutColor(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, Color: false})

	input := map[string]any{"name": "alice"}
	err := p.Write(input)
	require.NoError(t, err)

	assertNoColor(t, buf.String())
}

// YAML output tests
func TestWrite_YAML_SimpleObject(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, ToFormat: "yaml"})

	input := map[string]any{"name": "alice", "age": 30}
	err := p.Write(input)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name:")
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "age:")
	assert.Contains(t, output, "30")

	// Verify it's valid YAML
	var result map[string]any
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "alice", result["name"])
	assert.Equal(t, 30, result["age"])
}

func TestWrite_YAML_Array(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, ToFormat: "yaml"})

	input := []any{"alice", "bob", "charlie"}
	err := p.Write(input)
	require.NoError(t, err)

	var result []any
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "alice", result[0])
}

func TestWrite_YAML_NestedStructure(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, ToFormat: "yaml"})

	input := map[string]any{
		"user": map[string]any{
			"name":  "alice",
			"email": "alice@example.com",
		},
		"tags": []any{"golang", "cli"},
	}
	err := p.Write(input)
	require.NoError(t, err)

	var result map[string]any
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	user := result["user"].(map[string]any)
	assert.Equal(t, "alice", user["name"])

	tags := result["tags"].([]any)
	assert.Len(t, tags, 2)
}

func TestWrite_YAML_MultipleDocuments(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, ToFormat: "yaml"})

	err := p.Write(map[string]any{"first": 1})
	require.NoError(t, err)
	err = p.Write(map[string]any{"second": 2})
	require.NoError(t, err)

	output := buf.String()
	// Each YAML document should be present
	assert.Contains(t, output, "first:")
	assert.Contains(t, output, "second:")
}

func TestWrite_YAML_AllTypes(t *testing.T) {
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, ToFormat: "yaml"})

	input := map[string]any{
		"string":  "hello",
		"number":  42.5,
		"integer": 10,
		"boolean": true,
		"null":    nil,
		"array":   []any{1, 2, 3},
		"object":  map[string]any{"nested": "value"},
	}
	err := p.Write(input)
	require.NoError(t, err)

	var result map[string]any
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result["string"])
	assert.Equal(t, 42.5, result["number"])
	assert.Equal(t, true, result["boolean"])
	assert.Nil(t, result["null"])
}

// Colorization tests
func TestColorizeJSON_SimpleObject(t *testing.T) {
	input := []byte(`{"name":"alice"}`)
	output := string(colorizeJSON(input))

	assertHasColor(t, output)
	assertHasColorType(t, output, colKey, "key color")
	assertHasColorType(t, output, colStr, "string color")
}

func TestColorizeJSON_WithNumbers(t *testing.T) {
	input := []byte(`{"age":30}`)
	output := string(colorizeJSON(input))

	assertHasColorType(t, output, colNum, "number color")
}

func TestColorizeJSON_WithBoolean(t *testing.T) {
	input := []byte(`{"active":true}`)
	output := string(colorizeJSON(input))

	assertHasColorType(t, output, colBoolNil, "boolean color")
}

func TestColorizeJSON_WithNull(t *testing.T) {
	input := []byte(`{"value":null}`)
	output := string(colorizeJSON(input))

	assertHasColorType(t, output, colBoolNil, "null color")
}

func TestColorizeJSON_Array(t *testing.T) {
	input := []byte(`[1,2,3]`)
	output := string(colorizeJSON(input))

	assertHasColorType(t, output, colNum, "number color")
	assertHasColorType(t, output, colPunct, "punctuation color")
}

func TestColorizeJSON_NestedObject(t *testing.T) {
	input := []byte(`{"user":{"name":"alice"}}`)
	output := string(colorizeJSON(input))

	assertHasColorType(t, output, colKey, "key color")
	assertHasColorType(t, output, colStr, "string color")
}

func TestColorizeJSON_EmptyObject(t *testing.T) {
	input := []byte(`{}`)
	output := colorizeJSON(input)

	assert.NotEmpty(t, output)
}

func TestColorizeJSON_EmptyArray(t *testing.T) {
	input := []byte(`[]`)
	output := colorizeJSON(input)

	assert.NotEmpty(t, output)
}

func TestColorizeJSON_EscapedQuotes(t *testing.T) {
	input := []byte(`{"message":"Hello \"world\""}`)
	output := colorizeJSON(input)

	// Should handle escaped quotes properly
	assert.Contains(t, string(output), "Hello")
	assert.Contains(t, string(output), "world")
}

func TestColorizeJSON_EndsWithNewline(t *testing.T) {
	input := []byte(`{"key":"value"}`)
	output := colorizeJSON(input)

	assert.True(t, bytes.HasSuffix(output, []byte("\n")))
}

func TestColorizeJSON_PreservesNewline(t *testing.T) {
	input := []byte(`{"key":"value"}` + "\n")
	output := colorizeJSON(input)

	assert.True(t, bytes.HasSuffix(output, []byte("\n")))
}

func TestColorizeJSON_MultilineInput(t *testing.T) {
	input := []byte(`{
  "name": "alice",
  "age": 30
}`)
	output := string(colorizeJSON(input))

	// Should preserve formatting
	assert.Contains(t, output, "\n")
	assertHasColorType(t, output, colKey, "key color")
}

func TestColorizeJSON_ComplexNested(t *testing.T) {
	input := []byte(`{"users":[{"name":"alice","active":true},{"name":"bob","active":false}],"count":2}`)
	output := string(colorizeJSON(input))

	// Should handle complex nesting with all color types
	assertHasColorType(t, output, colKey, "key color")
	assertHasColorType(t, output, colStr, "string color")
	assertHasColorType(t, output, colBoolNil, "boolean color")
	assertHasColorType(t, output, colNum, "number color")
}

// Format conversion tests
func TestWrite_JSONToYAMLConversion(t *testing.T) {
	// This tests the ability to read JSON and output YAML
	buf := &bytes.Buffer{}
	p := New(Options{Writer: buf, ToFormat: "yaml"})

	input := map[string]any{
		"name":  "alice",
		"age":   30,
		"email": "alice@example.com",
	}
	err := p.Write(input)
	require.NoError(t, err)

	// Output should be valid YAML
	var result map[string]any
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, input["name"], result["name"])
	assert.Equal(t, input["age"], result["age"])
	assert.Equal(t, input["email"], result["email"])
}
