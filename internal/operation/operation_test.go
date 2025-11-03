package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for parsePath function
func TestParsePath_SimplePath(t *testing.T) {
	segs, err := parsePath("user")
	require.NoError(t, err)
	require.Len(t, segs, 1)
	assert.Equal(t, "user", segs[0].key)
	assert.Nil(t, segs[0].idx)
}

func TestParsePath_NestedPath(t *testing.T) {
	segs, err := parsePath("user.name")
	require.NoError(t, err)
	require.Len(t, segs, 2)
	assert.Equal(t, "user", segs[0].key)
	assert.Equal(t, "name", segs[1].key)
	assert.Nil(t, segs[0].idx)
	assert.Nil(t, segs[1].idx)
}

func TestParsePath_DeepNesting(t *testing.T) {
	segs, err := parsePath("a.b.c.d.e")
	require.NoError(t, err)
	require.Len(t, segs, 5)
	assert.Equal(t, "a", segs[0].key)
	assert.Equal(t, "e", segs[4].key)
}

func TestParsePath_WithArrayIndex(t *testing.T) {
	segs, err := parsePath("items[0]")
	require.NoError(t, err)
	require.Len(t, segs, 1)
	assert.Equal(t, "items", segs[0].key)
	require.NotNil(t, segs[0].idx)
	assert.Equal(t, 0, *segs[0].idx)
}

func TestParsePath_WithArrayIndexNonZero(t *testing.T) {
	segs, err := parsePath("items[42]")
	require.NoError(t, err)
	require.Len(t, segs, 1)
	assert.Equal(t, "items", segs[0].key)
	require.NotNil(t, segs[0].idx)
	assert.Equal(t, 42, *segs[0].idx)
}

func TestParsePath_WithWildcard(t *testing.T) {
	segs, err := parsePath("items[*]")
	require.NoError(t, err)
	require.Len(t, segs, 1)
	assert.Equal(t, "items", segs[0].key)
	require.NotNil(t, segs[0].idx)
	assert.Equal(t, -1, *segs[0].idx)
}

func TestParsePath_NestedWithArray(t *testing.T) {
	segs, err := parsePath("users[0].name")
	require.NoError(t, err)
	require.Len(t, segs, 2)
	assert.Equal(t, "users", segs[0].key)
	require.NotNil(t, segs[0].idx)
	assert.Equal(t, 0, *segs[0].idx)
	assert.Equal(t, "name", segs[1].key)
	assert.Nil(t, segs[1].idx)
}

func TestParsePath_ComplexNested(t *testing.T) {
	segs, err := parsePath("org.teams[5].members[2].profile.email")
	require.NoError(t, err)
	require.Len(t, segs, 5)

	assert.Equal(t, "org", segs[0].key)
	assert.Nil(t, segs[0].idx)

	assert.Equal(t, "teams", segs[1].key)
	require.NotNil(t, segs[1].idx)
	assert.Equal(t, 5, *segs[1].idx)

	assert.Equal(t, "members", segs[2].key)
	require.NotNil(t, segs[2].idx)
	assert.Equal(t, 2, *segs[2].idx)

	assert.Equal(t, "profile", segs[3].key)
	assert.Nil(t, segs[3].idx)

	assert.Equal(t, "email", segs[4].key)
	assert.Nil(t, segs[4].idx)
}

func TestParsePath_EmptyPath(t *testing.T) {
	_, err := parsePath("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestParsePath_InvalidBracketNoClose(t *testing.T) {
	_, err := parsePath("items[0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestParsePath_BracketNoOpen(t *testing.T) {
	// "items0]" is actually valid - it's just a key named "items0]"
	segs, err := parsePath("items0]")
	require.NoError(t, err)
	require.Len(t, segs, 1)
	assert.Equal(t, "items0]", segs[0].key)
}

func TestParsePath_InvalidEmptyBrackets(t *testing.T) {
	_, err := parsePath("items[]")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty index")
}

func TestParsePath_InvalidNegativeIndex(t *testing.T) {
	_, err := parsePath("items[-1]")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestParsePath_InvalidNonNumericIndex(t *testing.T) {
	_, err := parsePath("items[abc]")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestParsePath_InvalidBracketAtStart(t *testing.T) {
	_, err := parsePath("[0]")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestParsePath_MultipleWildcards(t *testing.T) {
	segs, err := parsePath("items[*].tags[*]")
	require.NoError(t, err)
	require.Len(t, segs, 2)

	assert.Equal(t, "items", segs[0].key)
	require.NotNil(t, segs[0].idx)
	assert.Equal(t, -1, *segs[0].idx)

	assert.Equal(t, "tags", segs[1].key)
	require.NotNil(t, segs[1].idx)
	assert.Equal(t, -1, *segs[1].idx)
}

// Tests for expandWildcardPaths function
func TestExpandWildcardPaths_NoWildcard(t *testing.T) {
	input := map[string]any{
		"name": "alice",
	}
	paths, err := expandWildcardPaths(input, "name")
	require.NoError(t, err)
	// No wildcards means it returns the path (since the path exists)
	require.Len(t, paths, 1)
	assert.Equal(t, "name", paths[0])
}

func TestExpandWildcardPaths_SimpleWildcard(t *testing.T) {
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	paths, err := expandWildcardPaths(input, "items[*]")
	require.NoError(t, err)
	require.Len(t, paths, 3)
	assert.Equal(t, "items[0]", paths[0])
	assert.Equal(t, "items[1]", paths[1])
	assert.Equal(t, "items[2]", paths[2])
}

func TestExpandWildcardPaths_WildcardWithNestedField(t *testing.T) {
	input := map[string]any{
		"users": []any{
			map[string]any{"name": "alice"},
			map[string]any{"name": "bob"},
		},
	}
	paths, err := expandWildcardPaths(input, "users[*].name")
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.Equal(t, "users[0].name", paths[0])
	assert.Equal(t, "users[1].name", paths[1])
}

func TestExpandWildcardPaths_EmptyArray(t *testing.T) {
	input := map[string]any{
		"items": []any{},
	}
	paths, err := expandWildcardPaths(input, "items[*]")
	require.NoError(t, err)
	assert.Empty(t, paths)
}

func TestExpandWildcardPaths_NonExistentPath(t *testing.T) {
	input := map[string]any{
		"name": "alice",
	}
	paths, err := expandWildcardPaths(input, "missing[*].field")
	require.NoError(t, err)
	assert.Empty(t, paths)
}

func TestExpandWildcardPaths_TypeMismatch(t *testing.T) {
	input := map[string]any{
		"items": "not an array",
	}
	paths, err := expandWildcardPaths(input, "items[*]")
	require.NoError(t, err)
	assert.Empty(t, paths)
}

func TestExpandWildcardPaths_MultipleWildcards(t *testing.T) {
	input := map[string]any{
		"teams": []any{
			map[string]any{
				"members": []any{"alice", "bob"},
			},
			map[string]any{
				"members": []any{"charlie", "diana", "eve"},
			},
		},
	}
	paths, err := expandWildcardPaths(input, "teams[*].members[*]")
	require.NoError(t, err)
	require.Len(t, paths, 5)
	assert.Equal(t, "teams[0].members[0]", paths[0])
	assert.Equal(t, "teams[0].members[1]", paths[1])
	assert.Equal(t, "teams[1].members[0]", paths[2])
	assert.Equal(t, "teams[1].members[1]", paths[3])
	assert.Equal(t, "teams[1].members[2]", paths[4])
}

func TestExpandWildcardPaths_DeepNesting(t *testing.T) {
	input := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"c": map[string]any{
						"d": "value",
					},
				},
			},
		},
	}
	paths, err := expandWildcardPaths(input, "a.b[*].c.d")
	require.NoError(t, err)
	require.Len(t, paths, 1)
	assert.Equal(t, "a.b[0].c.d", paths[0])
}

func TestExpandWildcardPaths_InvalidPath(t *testing.T) {
	input := map[string]any{"name": "alice"}
	_, err := expandWildcardPaths(input, "")
	assert.Error(t, err)
}

// Tests for buildPath helper function
func TestBuildPath_EmptyCurrent(t *testing.T) {
	result := buildPath("", "key")
	assert.Equal(t, "key", result)
}

func TestBuildPath_WithCurrent(t *testing.T) {
	result := buildPath("user", "name")
	assert.Equal(t, "user.name", result)
}

func TestBuildPath_Nested(t *testing.T) {
	result := buildPath("a.b.c", "d")
	assert.Equal(t, "a.b.c.d", result)
}

// Tests for asStringMap helper function
func TestAsStringMap_ValidMap(t *testing.T) {
	input := map[string]any{"key": "value"}
	result, ok := asStringMap(input)
	assert.True(t, ok)
	assert.Equal(t, input, result)
}

func TestAsStringMap_NotMap(t *testing.T) {
	input := "not a map"
	_, ok := asStringMap(input)
	assert.False(t, ok)
}

func TestAsStringMap_Nil(t *testing.T) {
	_, ok := asStringMap(nil)
	assert.False(t, ok)
}

// Tests for asSlice helper function
func TestAsSlice_ValidSlice(t *testing.T) {
	input := []any{1, 2, 3}
	result, ok := asSlice(input)
	assert.True(t, ok)
	assert.Equal(t, input, result)
}

func TestAsSlice_NotSlice(t *testing.T) {
	input := "not a slice"
	_, ok := asSlice(input)
	assert.False(t, ok)
}

func TestAsSlice_Nil(t *testing.T) {
	_, ok := asSlice(nil)
	assert.False(t, ok)
}

// Tests for getAtPath function
func TestGetAtPath_SimplePath(t *testing.T) {
	input := map[string]any{"name": "alice"}
	segs, _ := parsePath("name")
	val, ok := getAtPath(input, segs)
	assert.True(t, ok)
	assert.Equal(t, "alice", val)
}

func TestGetAtPath_NestedPath(t *testing.T) {
	input := map[string]any{
		"user": map[string]any{
			"name": "alice",
		},
	}
	segs, _ := parsePath("user.name")
	val, ok := getAtPath(input, segs)
	assert.True(t, ok)
	assert.Equal(t, "alice", val)
}

func TestGetAtPath_ArrayIndex(t *testing.T) {
	input := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	segs, _ := parsePath("items[1]")
	val, ok := getAtPath(input, segs)
	assert.True(t, ok)
	assert.Equal(t, "second", val)
}

func TestGetAtPath_NonExistentKey(t *testing.T) {
	input := map[string]any{"name": "alice"}
	segs, _ := parsePath("missing")
	_, ok := getAtPath(input, segs)
	assert.False(t, ok)
}

func TestGetAtPath_ArrayIndexOutOfBounds(t *testing.T) {
	input := map[string]any{
		"items": []any{1, 2, 3},
	}
	segs, _ := parsePath("items[10]")
	_, ok := getAtPath(input, segs)
	assert.False(t, ok)
}

func TestGetAtPath_TypeMismatch(t *testing.T) {
	input := map[string]any{
		"items": "not an array",
	}
	segs, _ := parsePath("items[0]")
	_, ok := getAtPath(input, segs)
	assert.False(t, ok)
}

func TestGetAtPath_EmptySegments(t *testing.T) {
	input := map[string]any{"name": "alice"}
	val, ok := getAtPath(input, []segment{})
	assert.True(t, ok)
	assert.Equal(t, input, val)
}

func TestGetAtPath_ComplexNested(t *testing.T) {
	input := map[string]any{
		"org": map[string]any{
			"teams": []any{
				map[string]any{
					"members": []any{
						map[string]any{"name": "alice"},
						map[string]any{"name": "bob"},
					},
				},
			},
		},
	}
	segs, _ := parsePath("org.teams[0].members[1].name")
	val, ok := getAtPath(input, segs)
	assert.True(t, ok)
	assert.Equal(t, "bob", val)
}
