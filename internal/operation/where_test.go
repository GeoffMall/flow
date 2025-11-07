package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhere_SingleCondition(t *testing.T) {
	where, err := NewWhere([]string{"name=Alice"})
	assert.NoError(t, err)

	// Match
	result, err := where.Apply(map[string]any{"name": "Alice", "age": 30})
	assert.NoError(t, err)
	assert.NotEqual(t, Filtered, result)

	// No match
	result, err = where.Apply(map[string]any{"name": "Bob", "age": 25})
	assert.NoError(t, err)
	assert.Equal(t, Filtered, result)
}

func TestWhere_MultipleConditions(t *testing.T) {
	where, err := NewWhere([]string{"name=Alice", "age=30"})
	assert.NoError(t, err)

	// Both match
	result, err := where.Apply(map[string]any{"name": "Alice", "age": 30})
	assert.NoError(t, err)
	assert.NotEqual(t, Filtered, result)

	// Only one matches
	result, err = where.Apply(map[string]any{"name": "Alice", "age": 25})
	assert.NoError(t, err)
	assert.Equal(t, Filtered, result)
}

func TestWhere_MissingField(t *testing.T) {
	where, err := NewWhere([]string{"name=Alice"})
	assert.NoError(t, err)

	// Field doesn't exist
	result, err := where.Apply(map[string]any{"age": 30})
	assert.NoError(t, err)
	assert.Equal(t, Filtered, result)
}

func TestWhere_EmptyConditions(t *testing.T) {
	where, err := NewWhere([]string{})
	assert.NoError(t, err)

	// No conditions, should pass through
	doc := map[string]any{"name": "Alice"}
	result, err := where.Apply(doc)
	assert.NoError(t, err)
	assert.Equal(t, doc, result)
}

func TestWhere_InvalidCondition(t *testing.T) {
	// Missing equals sign
	_, err := NewWhere([]string{"invalid"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be in format key=value")
}

func TestWhere_EmptyKey(t *testing.T) {
	_, err := NewWhere([]string{"=value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key cannot be empty")
}

func TestWhere_Description(t *testing.T) {
	where, err := NewWhere([]string{"name=Alice", "age=30"})
	assert.NoError(t, err)

	desc := where.Description()
	assert.Contains(t, desc, "name=Alice")
	assert.Contains(t, desc, "age=30")
	assert.Contains(t, desc, "AND")
}
