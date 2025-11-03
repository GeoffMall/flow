package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDelete_NonNil(t *testing.T) {
	d := NewDelete([]string{"key1", "key2"})
	assert.NotNil(t, d)
}

func TestDelete_SimpleKey(t *testing.T) {
	del := NewDelete([]string{"name"})
	input := map[string]any{
		"name": "alice",
		"age":  30,
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"age": 30}
	assert.Equal(t, expected, result)
}

func TestDelete_NestedPath(t *testing.T) {
	del := NewDelete([]string{"user.name"})
	input := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"age":  30,
		},
		"id": 123,
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{
			"age": 30,
		},
		"id": 123,
	}
	assert.Equal(t, expected, result)
}

func TestDelete_DeeplyNested(t *testing.T) {
	del := NewDelete([]string{"a.b.c.d"})
	input := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": "delete me",
					"e": "keep me",
				},
			},
		},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"e": "keep me",
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_ArrayElement(t *testing.T) {
	del := NewDelete([]string{"items[1]"})
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{"first", "third"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_ArrayElementFirst(t *testing.T) {
	del := NewDelete([]string{"items[0]"})
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{"second", "third"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_ArrayElementLast(t *testing.T) {
	del := NewDelete([]string{"items[2]"})
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{"first", "second"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_NestedArrayElement(t *testing.T) {
	del := NewDelete([]string{"users[1].email"})
	input := map[string]any{
		"users": []any{
			map[string]any{"name": "alice", "email": "alice@example.com"},
			map[string]any{"name": "bob", "email": "bob@example.com"},
			map[string]any{"name": "charlie", "email": "charlie@example.com"},
		},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"users": []any{
			map[string]any{"name": "alice", "email": "alice@example.com"},
			map[string]any{"name": "bob"},
			map[string]any{"name": "charlie", "email": "charlie@example.com"},
		},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_MultiplePaths(t *testing.T) {
	del := NewDelete([]string{"name", "email"})
	input := map[string]any{
		"name":  "alice",
		"email": "alice@example.com",
		"age":   30,
		"city":  "NYC",
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"age":  30,
		"city": "NYC",
	}
	assert.Equal(t, expected, result)
}

func TestDelete_NonExistentPath(t *testing.T) {
	del := NewDelete([]string{"missing"})
	input := map[string]any{"name": "alice"}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestDelete_NonExistentNestedPath(t *testing.T) {
	del := NewDelete([]string{"user.missing"})
	input := map[string]any{
		"user": map[string]any{"name": "alice"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{"name": "alice"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_PathDoesNotExist(t *testing.T) {
	del := NewDelete([]string{"user.profile.email"})
	input := map[string]any{
		"user": map[string]any{"name": "alice"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{"name": "alice"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_ArrayIndexOutOfBounds(t *testing.T) {
	del := NewDelete([]string{"items[10]"})
	input := map[string]any{
		"items": []any{1, 2, 3},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{1, 2, 3},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_TypeMismatch_ArrayAsMap(t *testing.T) {
	del := NewDelete([]string{"items.name"})
	input := map[string]any{
		"items": []any{1, 2, 3},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{1, 2, 3},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_TypeMismatch_MapAsArray(t *testing.T) {
	del := NewDelete([]string{"user[0]"})
	input := map[string]any{
		"user": map[string]any{"name": "alice"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{"name": "alice"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_ScalarInput(t *testing.T) {
	del := NewDelete([]string{"name"})
	input := "just a string"
	result, err := del.Apply(input)
	require.NoError(t, err)

	assert.Equal(t, "just a string", result)
}

func TestDelete_NilInput(t *testing.T) {
	del := NewDelete([]string{"name"})
	result, err := del.Apply(nil)
	require.NoError(t, err)

	assert.Nil(t, result)
}

func TestDelete_EmptyPaths(t *testing.T) {
	del := NewDelete([]string{})
	input := map[string]any{"name": "alice", "age": 30}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice", "age": 30}
	assert.Equal(t, expected, result)
}

func TestDelete_InvalidPath_Empty(t *testing.T) {
	del := NewDelete([]string{""})
	input := map[string]any{"name": "alice"}
	_, err := del.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestDelete_InvalidPath_BadArraySyntax(t *testing.T) {
	del := NewDelete([]string{"items[abc]"})
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := del.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestDelete_Description(t *testing.T) {
	del := NewDelete([]string{"name", "email", "age"})
	desc := del.Description()
	assert.Equal(t, "delete(name, email, age)", desc)
}

func TestDelete_WildcardExpansion(t *testing.T) {
	del := NewDelete([]string{"items[*].secret"})
	input := map[string]any{
		"items": []any{
			map[string]any{"name": "alice", "secret": "password1"},
			map[string]any{"name": "bob", "secret": "password2"},
			map[string]any{"name": "charlie", "secret": "password3"},
		},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{
			map[string]any{"name": "alice"},
			map[string]any{"name": "bob"},
			map[string]any{"name": "charlie"},
		},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_ComplexNestedWithArray(t *testing.T) {
	del := NewDelete([]string{"org.teams[0].members[1]"})
	input := map[string]any{
		"org": map[string]any{
			"teams": []any{
				map[string]any{
					"members": []any{
						map[string]any{"name": "alice"},
						map[string]any{"name": "bob"},
						map[string]any{"name": "charlie"},
					},
				},
			},
		},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"org": map[string]any{
			"teams": []any{
				map[string]any{
					"members": []any{
						map[string]any{"name": "alice"},
						map[string]any{"name": "charlie"},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_SequentialArrayDeletions(t *testing.T) {
	// Deleting multiple elements from the same array
	// Note: indices shift after each deletion, so order matters
	del := NewDelete([]string{"items[0]", "items[0]"})
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{"third"},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_EntireArray(t *testing.T) {
	del := NewDelete([]string{"items"})
	input := map[string]any{
		"items": []any{1, 2, 3},
		"name":  "test",
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "test"}
	assert.Equal(t, expected, result)
}

func TestDelete_EmptyArray(t *testing.T) {
	del := NewDelete([]string{"items[0]"})
	input := map[string]any{
		"items": []any{},
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{},
	}
	assert.Equal(t, expected, result)
}

func TestDelete_PreservesOtherFields(t *testing.T) {
	del := NewDelete([]string{"user.password"})
	input := map[string]any{
		"user": map[string]any{
			"name":     "alice",
			"email":    "alice@example.com",
			"password": "secret",
			"age":      30,
		},
		"id": 123,
	}
	result, err := del.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{
			"name":  "alice",
			"email": "alice@example.com",
			"age":   30,
		},
		"id": 123,
	}
	assert.Equal(t, expected, result)
}
