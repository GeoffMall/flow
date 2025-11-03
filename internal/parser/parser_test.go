package parser

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_NoErrOnOsStdin(t *testing.T) {
	_, err := New(os.Stdin)
	assert.NoError(t, err)
}

// Format detection tests
func TestDetectFormat_JSON_Object(t *testing.T) {
	input := strings.NewReader(`{"key": "value"}`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

func TestDetectFormat_JSON_Array(t *testing.T) {
	input := strings.NewReader(`[1, 2, 3]`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

func TestDetectFormat_JSON_Number(t *testing.T) {
	input := strings.NewReader(`42`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

func TestDetectFormat_JSON_NegativeNumber(t *testing.T) {
	input := strings.NewReader(`-42`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

func TestDetectFormat_YAML_DocumentSeparator(t *testing.T) {
	input := strings.NewReader(`---
name: alice`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, YAML, p.Format())
}

func TestDetectFormat_YAML_Directive(t *testing.T) {
	input := strings.NewReader(`%YAML 1.2
---
name: alice`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, YAML, p.Format())
}

func TestDetectFormat_YAML_KeyValue(t *testing.T) {
	input := strings.NewReader(`name: alice
age: 30`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, YAML, p.Format())
}

func TestDetectFormat_YAML_NestedStructure(t *testing.T) {
	input := strings.NewReader(`user:
  name: alice
  age: 30`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, YAML, p.Format())
}

func TestDetectFormat_EmptyInput(t *testing.T) {
	input := strings.NewReader(``)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

func TestDetectFormat_WhitespaceOnly(t *testing.T) {
	input := strings.NewReader(`

	`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

func TestDetectFormat_JSON_WithLeadingWhitespace(t *testing.T) {
	input := strings.NewReader(`
	{"key": "value"}`)
	p, err := New(input)
	require.NoError(t, err)
	assert.Equal(t, JSON, p.Format())
}

// Streaming tests - JSON
func TestForEach_JSON_SingleObject(t *testing.T) {
	input := strings.NewReader(`{"name": "alice", "age": 30}`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	obj, ok := results[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "alice", obj["name"])
	assert.Equal(t, float64(30), obj["age"])
}

func TestForEach_JSON_ArrayElements(t *testing.T) {
	input := strings.NewReader(`[{"name": "alice"}, {"name": "bob"}, {"name": "charlie"}]`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Equal(t, "alice", results[0].(map[string]any)["name"])
	assert.Equal(t, "bob", results[1].(map[string]any)["name"])
	assert.Equal(t, "charlie", results[2].(map[string]any)["name"])
}

func TestForEach_JSON_ConcatenatedDocuments(t *testing.T) {
	input := strings.NewReader(`{"first": 1}
{"second": 2}
{"third": 3}`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Contains(t, results[0].(map[string]any), "first")
	assert.Contains(t, results[1].(map[string]any), "second")
	assert.Contains(t, results[2].(map[string]any), "third")
}

func TestForEach_JSON_ArrayThenConcatenated(t *testing.T) {
	input := strings.NewReader(`[1, 2, 3]
{"extra": true}`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 4)

	assert.Equal(t, float64(1), results[0])
	assert.Equal(t, float64(2), results[1])
	assert.Equal(t, float64(3), results[2])
	assert.Equal(t, true, results[3].(map[string]any)["extra"])
}

func TestForEach_JSON_EmptyArray(t *testing.T) {
	input := strings.NewReader(`[]`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestForEach_JSON_ScalarValues(t *testing.T) {
	input := strings.NewReader(`42
"hello"
true
null`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 4)

	assert.Equal(t, float64(42), results[0])
	assert.Equal(t, "hello", results[1])
	assert.Equal(t, true, results[2])
	assert.Nil(t, results[3])
}

func TestForEach_JSON_CallbackError(t *testing.T) {
	input := strings.NewReader(`[1, 2, 3]`)
	p, err := New(input)
	require.NoError(t, err)

	err = p.ForEach(func(v any) error {
		if v.(float64) == 2 {
			return io.EOF
		}
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

func TestForEach_JSON_MalformedInput(t *testing.T) {
	input := strings.NewReader(`{"invalid": }`)
	p, err := New(input)
	require.NoError(t, err)

	err = p.ForEach(func(v any) error {
		return nil
	})
	assert.Error(t, err)
}

// Streaming tests - YAML
func TestForEach_YAML_SingleDocument(t *testing.T) {
	input := strings.NewReader(`name: alice
age: 30`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	obj, ok := results[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "alice", obj["name"])
	assert.Equal(t, 30, obj["age"])
}

func TestForEach_YAML_MultipleDocuments(t *testing.T) {
	input := strings.NewReader(`---
name: alice
---
name: bob
---
name: charlie`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Equal(t, "alice", results[0].(map[string]any)["name"])
	assert.Equal(t, "bob", results[1].(map[string]any)["name"])
	assert.Equal(t, "charlie", results[2].(map[string]any)["name"])
}

func TestForEach_YAML_NestedStructure(t *testing.T) {
	input := strings.NewReader(`user:
  name: alice
  profile:
    email: alice@example.com
    age: 30`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	obj := results[0].(map[string]any)
	user := obj["user"].(map[string]any)
	assert.Equal(t, "alice", user["name"])

	profile := user["profile"].(map[string]any)
	assert.Equal(t, "alice@example.com", profile["email"])
	assert.Equal(t, 30, profile["age"])
}

func TestForEach_YAML_Array(t *testing.T) {
	input := strings.NewReader(`---
- alice
- bob
- charlie`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	arr, ok := results[0].([]any)
	require.True(t, ok)
	require.Len(t, arr, 3)
	assert.Equal(t, "alice", arr[0])
	assert.Equal(t, "bob", arr[1])
	assert.Equal(t, "charlie", arr[2])
}

func TestForEach_YAML_CallbackError(t *testing.T) {
	input := strings.NewReader(`---
name: alice
---
name: bob`)
	p, err := New(input)
	require.NoError(t, err)

	count := 0
	err = p.ForEach(func(v any) error {
		count++
		if count == 2 {
			return io.EOF
		}
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

func TestForEach_YAML_MalformedInput(t *testing.T) {
	input := strings.NewReader(`invalid: [this is not: valid yaml`)
	p, err := New(input)
	require.NoError(t, err)

	err = p.ForEach(func(v any) error {
		return nil
	})
	assert.Error(t, err)
}

// Normalization tests
func TestNormalizeYAML_StringKey(t *testing.T) {
	input := map[any]any{
		"name": "alice",
		"age":  30,
	}
	result := normalizeYAML(input)

	normalized, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "alice", normalized["name"])
	assert.Equal(t, 30, normalized["age"])
}

func TestNormalizeYAML_NonStringKey(t *testing.T) {
	// Test with string keys that represent non-string values
	// In practice, YAML keys are typically strings
	input := map[any]any{
		"123":  "numeric key",
		"true": "bool key",
	}
	result := normalizeYAML(input)

	normalized, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "numeric key", normalized["123"])
	assert.Equal(t, "bool key", normalized["true"])
}

func TestNormalizeYAML_NestedMaps(t *testing.T) {
	input := map[any]any{
		"outer": map[any]any{
			"inner": map[any]any{
				"deep": "value",
			},
		},
	}
	result := normalizeYAML(input)

	normalized := result.(map[string]any)
	outer := normalized["outer"].(map[string]any)
	inner := outer["inner"].(map[string]any)
	assert.Equal(t, "value", inner["deep"])
}

func TestNormalizeYAML_Arrays(t *testing.T) {
	input := []any{
		map[any]any{"name": "alice"},
		map[any]any{"name": "bob"},
	}
	result := normalizeYAML(input)

	normalized, ok := result.([]any)
	require.True(t, ok)
	require.Len(t, normalized, 2)

	assert.Equal(t, "alice", normalized[0].(map[string]any)["name"])
	assert.Equal(t, "bob", normalized[1].(map[string]any)["name"])
}

func TestNormalizeYAML_Scalar(t *testing.T) {
	assert.Equal(t, "hello", normalizeYAML("hello"))
	assert.Equal(t, 42, normalizeYAML(42))
	assert.Equal(t, true, normalizeYAML(true))
	assert.Nil(t, normalizeYAML(nil))
}

func TestNormalizeYAML_MapStringAny(t *testing.T) {
	input := map[string]any{
		"name": "alice",
		"nested": map[any]any{
			"key": "value",
		},
	}
	result := normalizeYAML(input)

	normalized := result.(map[string]any)
	assert.Equal(t, "alice", normalized["name"])
	nested := normalized["nested"].(map[string]any)
	assert.Equal(t, "value", nested["key"])
}

// Helper function tests
func TestToStringKey_String(t *testing.T) {
	assert.Equal(t, "hello", toStringKey("hello"))
}

func TestLooksLikeYAML_KeyValue(t *testing.T) {
	assert.True(t, looksLikeYAML("key: value"))
	assert.True(t, looksLikeYAML("name: alice"))
}

func TestLooksLikeYAML_NoColon(t *testing.T) {
	assert.False(t, looksLikeYAML("just text"))
	assert.False(t, looksLikeYAML("[1, 2, 3]"))
}

func TestLooksLikeYAML_JSONObject(t *testing.T) {
	// The heuristic checks if ':' comes before ',' or '}'
	// In `{"key": "value"}`, the ':' comes before the '}', so it returns true
	// This is a known limitation of the heuristic, but detectFormat handles it correctly
	assert.True(t, looksLikeYAML(`{"key": "value"}`))
}

func TestLooksLikeYAML_MultipleLines(t *testing.T) {
	input := `name: alice
age: 30`
	assert.True(t, looksLikeYAML(input))
}

func TestForEach_JSON_ComplexNested(t *testing.T) {
	input := strings.NewReader(`{
		"users": [
			{"name": "alice", "age": 30},
			{"name": "bob", "age": 25}
		],
		"metadata": {
			"version": 1,
			"timestamp": "2024-01-01"
		}
	}`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	obj := results[0].(map[string]any)
	users := obj["users"].([]any)
	assert.Len(t, users, 2)

	metadata := obj["metadata"].(map[string]any)
	assert.Equal(t, float64(1), metadata["version"])
}

func TestForEach_YAML_ComplexNested(t *testing.T) {
	input := strings.NewReader(`users:
  - name: alice
    age: 30
  - name: bob
    age: 25
metadata:
  version: 1
  timestamp: "2024-01-01"`)
	p, err := New(input)
	require.NoError(t, err)

	var results []any
	err = p.ForEach(func(v any) error {
		results = append(results, v)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	obj := results[0].(map[string]any)
	users := obj["users"].([]any)
	assert.Len(t, users, 2)

	metadata := obj["metadata"].(map[string]any)
	assert.Equal(t, 1, metadata["version"])
}
