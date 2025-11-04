package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPick_EmptyPaths(t *testing.T) {
	// Empty paths should return input as-is in both modes
	pick := NewPick([]string{}, false)
	input := map[string]any{"user": "alice", "id": 123}
	result, err := pick.Apply(input)
	require.NoError(t, err)
	assert.Equal(t, input, result)
}

func TestPick_SimplePath(t *testing.T) {
	input := map[string]any{"name": "alice", "age": 30}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Equal(t, "alice", result) // JUST THE VALUE
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{"name": "alice"}
		assert.Equal(t, expected, result)
	})
}

func TestPick_NestedPath(t *testing.T) {
	input := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"age":  30,
		},
		"id": 123,
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"user.name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Equal(t, "alice", result) // JUST THE VALUE
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"user.name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{
			"user": map[string]any{
				"name": "alice",
			},
		}
		assert.Equal(t, expected, result)
	})
}

func TestPick_DeeplyNestedPath(t *testing.T) {
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

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"a.b.c.d.e"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Equal(t, "value", result) // JUST THE VALUE
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"a.b.c.d.e"}, true)
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
	})
}

func TestPick_ArrayIndex(t *testing.T) {
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items[0]"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Equal(t, "first", result) // JUST THE VALUE
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items[0]"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{
			"items": []any{"first"},
		}
		assert.Equal(t, expected, result)
	})
}

func TestPick_ArrayIndexNested(t *testing.T) {
	input := map[string]any{
		"items": []any{
			map[string]any{"name": "first", "id": 1},
			map[string]any{"name": "second", "id": 2},
			map[string]any{"name": "third", "id": 3},
		},
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items[1].name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Equal(t, "second", result) // JUST THE VALUE
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items[1].name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{
			"items": []any{
				nil,
				map[string]any{"name": "second"},
			},
		}
		assert.Equal(t, expected, result)
	})
}

func TestPick_WildcardExpansion(t *testing.T) {
	input := map[string]any{
		"items": []any{
			map[string]any{"name": "alice", "age": 30},
			map[string]any{"name": "bob", "age": 25},
			map[string]any{"name": "charlie", "age": 35},
		},
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items[*].name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Wildcard in single-pick mode returns array of values
		expected := []any{"alice", "bob", "charlie"}
		assert.Equal(t, expected, result)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items[*].name"}, true)
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
	})
}

func TestPick_WildcardMultipleFields(t *testing.T) {
	input := map[string]any{
		"users": []any{
			map[string]any{"name": "alice", "email": "alice@example.com", "age": 30},
			map[string]any{"name": "bob", "email": "bob@example.com", "age": 25},
		},
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"users[*].name", "users[*].email"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Multiple wildcards in multi-pick mode returns object with arrays
		expected := map[string]any{
			"name":  []any{"alice", "bob"},
			"email": []any{"alice@example.com", "bob@example.com"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"users[*].name", "users[*].email"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{
			"users": []any{
				map[string]any{"name": "alice", "email": "alice@example.com"},
				map[string]any{"name": "bob", "email": "bob@example.com"},
			},
		}
		assert.Equal(t, expected, result)
	})
}

func TestPick_MultiplePaths(t *testing.T) {
	input := map[string]any{
		"name":  "alice",
		"email": "alice@example.com",
		"age":   30,
		"city":  "NYC",
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"name", "email"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Multiple paths returns flattened object (same as hierarchy in this case)
		expected := map[string]any{
			"name":  "alice",
			"email": "alice@example.com",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"name", "email"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{
			"name":  "alice",
			"email": "alice@example.com",
		}
		assert.Equal(t, expected, result)
	})
}

func TestPick_NonExistentPath(t *testing.T) {
	input := map[string]any{"name": "alice"}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"missing.path"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil for missing paths (jq behavior)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"missing.path"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_MixedExistingAndNonExisting(t *testing.T) {
	input := map[string]any{
		"name": "alice",
		"age":  30,
	}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"name", "missing", "age"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Missing paths are skipped in multi-pick
		expected := map[string]any{
			"name": "alice",
			"age":  30,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"name", "missing", "age"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{
			"name": "alice",
			"age":  30,
		}
		assert.Equal(t, expected, result)
	})
}

func TestPick_InvalidPath_Empty(t *testing.T) {
	// Error tests - behavior is the same in both modes
	pick := NewPick([]string{""}, false)
	input := map[string]any{"name": "alice"}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestPick_InvalidPath_BadArraySyntax(t *testing.T) {
	pick := NewPick([]string{"items[abc]"}, false)
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPick_InvalidPath_EmptyBrackets(t *testing.T) {
	pick := NewPick([]string{"items[]"}, false)
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty index")
}

func TestPick_InvalidPath_MissingCloseBracket(t *testing.T) {
	pick := NewPick([]string{"items[0"}, false)
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPick_InvalidPath_NegativeIndex(t *testing.T) {
	pick := NewPick([]string{"items[-1]"}, false)
	input := map[string]any{"items": []any{1, 2, 3}}
	_, err := pick.Apply(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPick_ArrayIndexOutOfBounds(t *testing.T) {
	input := map[string]any{"items": []any{1, 2, 3}}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items[10]"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil for missing index
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items[10]"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_TypeMismatch_ArrayAsMap(t *testing.T) {
	input := map[string]any{"items": []any{1, 2, 3}}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items.name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil for type mismatch
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items.name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_TypeMismatch_MapAsArray(t *testing.T) {
	input := map[string]any{"user": map[string]any{"name": "alice"}}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"user[0]"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil for type mismatch
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"user[0]"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_ScalarInput(t *testing.T) {
	input := "just a string"

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil for scalar input
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_NilInput(t *testing.T) {
	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"name"}, false)
		result, err := pick.Apply(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil for nil input
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"name"}, true)
		result, err := pick.Apply(nil)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_Description(t *testing.T) {
	pick := NewPick([]string{"name", "email", "age"}, false)
	desc := pick.Description()
	assert.Equal(t, "pick(name, email, age)", desc)
}

func TestPick_EmptyWildcardArray(t *testing.T) {
	input := map[string]any{"items": []any{}}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items[*].name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Empty wildcard returns empty array
		assert.Equal(t, []any{}, result)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items[*].name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Empty array expands to no paths, so nothing is picked
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_WildcardNonArray(t *testing.T) {
	input := map[string]any{"items": "not an array"}

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"items[*].name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		assert.Equal(t, []any{}, result) // Empty array for non-array wildcard
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"items[*].name"}, true)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		expected := map[string]any{}
		assert.Equal(t, expected, result)
	})
}

func TestPick_ComplexNestedWithWildcard(t *testing.T) {
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

	t.Run("jq-like", func(t *testing.T) {
		pick := NewPick([]string{"org.teams[*].members[0].name"}, false)
		result, err := pick.Apply(input)
		require.NoError(t, err)
		// Returns array of values
		expected := []any{"alice", "charlie"}
		assert.Equal(t, expected, result)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		pick := NewPick([]string{"org.teams[*].members[0].name"}, true)
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
	})
}

func TestPick_AllTypes(t *testing.T) {
	input := map[string]any{
		"string":  "hello",
		"number":  42.5,
		"boolean": true,
		"null":    nil,
		"array":   []any{1, 2, 3},
		"object":  map[string]any{"nested": "value"},
		"ignored": "this should not appear",
	}

	// Both modes have same behavior for multiple root-level picks
	pick := NewPick([]string{"string", "number", "boolean", "null", "array", "object"}, false)
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
