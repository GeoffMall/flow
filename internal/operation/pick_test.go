package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPick_EmptyPaths(t *testing.T) {
	pick := NewPick([]string{})
	input := map[string]any{"user": "alice", "id": 123}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	assert.Equal(t, input, result)
}

func TestPick_SimplePath(t *testing.T) {
	pick := NewPick([]string{"name"})
	input := map[string]any{"name": "alice", "age": 30}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestPick_NestedPath(t *testing.T) {
	pick := NewPick([]string{"user.name"})
	input := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"age":  30,
		},
		"id": 123,
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"user": map[string]any{
			"name": "alice",
		},
	}
	assert.Equal(t, expected, result)
}

func TestPick_DeeplyNestedPath(t *testing.T) {
	pick := NewPick([]string{"a.b.c.d.e"})
	input := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": "value",
						"f": "ignored",
					},
				},
			},
		},
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": "value",
					},
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestPick_ArrayIndex(t *testing.T) {
	pick := NewPick([]string{"items[0]"})
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"items": []any{"first"},
	}
	assert.Equal(t, expected, result)
}

func TestPick_ArrayIndexNested(t *testing.T) {
	pick := NewPick([]string{"items[1].name"})
	input := map[string]any{
		"items": []any{
			map[string]any{"name": "first", "id": 1},
			map[string]any{"name": "second", "id": 2},
			map[string]any{"name": "third", "id": 3},
		},
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"items": []any{
			nil,
			map[string]any{"name": "second"},
		},
	}
	assert.Equal(t, expected, result)
}

func TestPick_WildcardExpansion(t *testing.T) {
	pick := NewPick([]string{"items[*].name"})
	input := map[string]any{
		"items": []any{
			map[string]any{"name": "alice", "age": 30},
			map[string]any{"name": "bob", "age": 25},
			map[string]any{"name": "charlie", "age": 35},
		},
	}
	result, err := pick.Apply(input)
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

func TestPick_WildcardMultipleFields(t *testing.T) {
	pick := NewPick([]string{"users[*].name", "users[*].email"})
	input := map[string]any{
		"users": []any{
			map[string]any{"name": "alice", "email": "alice@example.com", "age": 30},
			map[string]any{"name": "bob", "email": "bob@example.com", "age": 25},
		},
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"users": []any{
			map[string]any{"name": "alice", "email": "alice@example.com"},
			map[string]any{"name": "bob", "email": "bob@example.com"},
		},
	}
	assert.Equal(t, expected, result)
}

func TestPick_MultiplePaths(t *testing.T) {
	pick := NewPick([]string{"name", "email"})
	input := map[string]any{
		"name":  "alice",
		"email": "alice@example.com",
		"age":   30,
		"city":  "NYC",
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"name":  "alice",
		"email": "alice@example.com",
	}
	assert.Equal(t, expected, result)
}

func TestPick_NonExistentPath(t *testing.T) {
	pick := NewPick([]string{"missing.path"})
	input := map[string]any{"name": "alice"}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_MixedExistingAndNonExisting(t *testing.T) {
	pick := NewPick([]string{"name", "missing", "age"})
	input := map[string]any{
		"name": "alice",
		"age":  30,
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"name": "alice",
		"age":  30,
	}
	assert.Equal(t, expected, result)
}

func TestPick_InvalidPath_Empty(t *testing.T) {
	pick := NewPick([]string{""})
	input := map[string]any{"name": "alice"}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestPick_InvalidPath_BadArraySyntax(t *testing.T) {
	pick := NewPick([]string{"items[abc]"})
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPick_InvalidPath_EmptyBrackets(t *testing.T) {
	pick := NewPick([]string{"items[]"})
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty index")
}

func TestPick_InvalidPath_MissingCloseBracket(t *testing.T) {
	pick := NewPick([]string{"items[0"})
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPick_InvalidPath_NegativeIndex(t *testing.T) {
	pick := NewPick([]string{"items[-1]"})
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPick_ArrayIndexOutOfBounds(t *testing.T) {
	pick := NewPick([]string{"items[10]"})
	input := map[string]any{"items": []any{1, 2, 3}}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_TypeMismatch_ArrayAsMap(t *testing.T) {
	pick := NewPick([]string{"items.name"})
	input := map[string]any{"items": []any{1, 2, 3}}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_TypeMismatch_MapAsArray(t *testing.T) {
	pick := NewPick([]string{"user[0]"})
	input := map[string]any{"user": map[string]any{"name": "alice"}}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_ScalarInput(t *testing.T) {
	pick := NewPick([]string{"name"})
	input := "just a string"
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_NilInput(t *testing.T) {
	pick := NewPick([]string{"name"})
	result, err := pick.Apply(nil)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_Description(t *testing.T) {
	pick := NewPick([]string{"name", "email", "age"})
	desc := pick.Description()
	assert.Equal(t, "pick(name, email, age)", desc)
}

func TestPick_EmptyWildcardArray(t *testing.T) {
	pick := NewPick([]string{"items[*].name"})
	input := map[string]any{"items": []any{}}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	// Empty array expands to no paths, so nothing is picked
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_WildcardNonArray(t *testing.T) {
	pick := NewPick([]string{"items[*].name"})
	input := map[string]any{"items": "not an array"}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{}
	assert.Equal(t, expected, result)
}

func TestPick_ComplexNestedWithWildcard(t *testing.T) {
	pick := NewPick([]string{"org.teams[*].members[0].name"})
	input := map[string]any{
		"org": map[string]any{
			"teams": []any{
				map[string]any{
					"members": []any{
						map[string]any{"name": "alice", "role": "lead"},
						map[string]any{"name": "bob", "role": "dev"},
					},
				},
				map[string]any{
					"members": []any{
						map[string]any{"name": "charlie", "role": "lead"},
						map[string]any{"name": "diana", "role": "dev"},
					},
				},
			},
		},
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"org": map[string]any{
			"teams": []any{
				map[string]any{
					"members": []any{
						map[string]any{"name": "alice"},
					},
				},
				map[string]any{
					"members": []any{
						map[string]any{"name": "charlie"},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestPick_AllTypes(t *testing.T) {
	pick := NewPick([]string{"string", "number", "boolean", "null", "array", "object"})
	input := map[string]any{
		"string":  "hello",
		"number":  42.5,
		"boolean": true,
		"null":    nil,
		"array":   []any{1, 2, 3},
		"object":  map[string]any{"nested": "value"},
		"ignored": "this should not appear",
	}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	expected := map[string]any{
		"string":  "hello",
		"number":  42.5,
		"boolean": true,
		"null":    nil,
		"array":   []any{1, 2, 3},
		"object":  map[string]any{"nested": "value"},
	}
	assert.Equal(t, expected, result)
}
