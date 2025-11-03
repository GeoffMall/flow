package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_SimplePath(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestSet_NestedPath(t *testing.T) {
	set, err := NewSetFromPairs([]string{"user.name=alice"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{
			"name": "alice",
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_DeeplyNested(t *testing.T) {
	set, err := NewSetFromPairs([]string{"a.b.c.d=value"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": "value",
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_OverwriteExisting(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=bob"})
	require.NoError(t, err)

	input := map[string]any{"name": "alice", "age": 30}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "bob", "age": 30}
	assert.Equal(t, expected, result)
}

func TestSet_NumberValue(t *testing.T) {
	set, err := NewSetFromPairs([]string{"age=42"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"age": float64(42)}
	assert.Equal(t, expected, result)
}

func TestSet_FloatValue(t *testing.T) {
	set, err := NewSetFromPairs([]string{"price=19.99"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"price": 19.99}
	assert.Equal(t, expected, result)
}

func TestSet_BooleanTrue(t *testing.T) {
	set, err := NewSetFromPairs([]string{"active=true"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"active": true}
	assert.Equal(t, expected, result)
}

func TestSet_BooleanFalse(t *testing.T) {
	set, err := NewSetFromPairs([]string{"active=false"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"active": false}
	assert.Equal(t, expected, result)
}

func TestSet_NullValue(t *testing.T) {
	set, err := NewSetFromPairs([]string{"value=null"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"value": nil}
	assert.Equal(t, expected, result)
}

func TestSet_JSONObject(t *testing.T) {
	set, err := NewSetFromPairs([]string{`user={"name":"alice","age":30}`})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"age":  float64(30),
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_JSONArray(t *testing.T) {
	set, err := NewSetFromPairs([]string{`tags=["go","cli","json"]`})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"tags": []any{"go", "cli", "json"},
	}
	assert.Equal(t, expected, result)
}

func TestSet_QuotedString(t *testing.T) {
	set, err := NewSetFromPairs([]string{`message="hello world"`})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"message": "hello world"}
	assert.Equal(t, expected, result)
}

func TestSet_UnquotedStringFallback(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestSet_ArrayIndex(t *testing.T) {
	set, err := NewSetFromPairs([]string{"items[0]=first"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{"first"},
	}
	assert.Equal(t, expected, result)
}

func TestSet_ArrayIndexNonZero(t *testing.T) {
	set, err := NewSetFromPairs([]string{"items[2]=third"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{nil, nil, "third"},
	}
	assert.Equal(t, expected, result)
}

func TestSet_ArrayIndexNested(t *testing.T) {
	set, err := NewSetFromPairs([]string{"items[1].name=bob"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{
			nil,
			map[string]any{"name": "bob"},
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_OverwriteScalarWithMap(t *testing.T) {
	set, err := NewSetFromPairs([]string{"user.name=alice"})
	require.NoError(t, err)

	input := map[string]any{"user": "old scalar value"}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"user": map[string]any{
			"name": "alice",
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_OverwriteArrayWithMap(t *testing.T) {
	set, err := NewSetFromPairs([]string{"data.field=value"})
	require.NoError(t, err)

	input := map[string]any{"data": []any{1, 2, 3}}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"data": map[string]any{
			"field": "value",
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_MultipleAssignments(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice", "age=30", "city=NYC"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"name": "alice",
		"age":  float64(30),
		"city": "NYC",
	}
	assert.Equal(t, expected, result)
}

func TestSet_MixedNestedAndFlat(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice", "profile.email=alice@example.com", "age=30"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"name": "alice",
		"age":  float64(30),
		"profile": map[string]any{
			"email": "alice@example.com",
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_PreservesUnmodifiedFields(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=bob"})
	require.NoError(t, err)

	input := map[string]any{
		"name":  "alice",
		"email": "alice@example.com",
		"age":   30,
	}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"name":  "bob",
		"email": "alice@example.com",
		"age":   30,
	}
	assert.Equal(t, expected, result)
}

func TestSet_NilInput(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice"})
	require.NoError(t, err)

	result, err := set.Apply(nil)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestSet_ScalarInput(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice"})
	require.NoError(t, err)

	result, err := set.Apply("scalar value")
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestSet_InvalidPair_NoEquals(t *testing.T) {
	_, err := NewSetFromPairs([]string{"invalid"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected path=value")
}

func TestSet_InvalidPair_EmptyPath(t *testing.T) {
	_, err := NewSetFromPairs([]string{"=value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestSet_InvalidPair_EmptyPathWithSpaces(t *testing.T) {
	_, err := NewSetFromPairs([]string{"  =value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestSet_EmptyValue(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name="})
	require.NoError(t, err)

	result, err := set.Apply(map[string]any{})
	require.NoError(t, err)

	expected := map[string]any{"name": ""}
	assert.Equal(t, expected, result)
}

func TestSet_ValueWithEquals(t *testing.T) {
	set, err := NewSetFromPairs([]string{"formula=a=b+c"})
	require.NoError(t, err)

	result, err := set.Apply(map[string]any{})
	require.NoError(t, err)

	expected := map[string]any{"formula": "a=b+c"}
	assert.Equal(t, expected, result)
}

func TestSet_Description(t *testing.T) {
	set, err := NewSetFromPairs([]string{"name=alice", "age=30", "city=NYC"})
	require.NoError(t, err)

	desc := set.Description()
	assert.Equal(t, "set(name, age, city)", desc)
}

func TestSet_ExpandArrayToAccommodateIndex(t *testing.T) {
	set, err := NewSetFromPairs([]string{"items[5]=sixth"})
	require.NoError(t, err)

	input := map[string]any{"items": []any{"first", "second"}}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"items": []any{"first", "second", nil, nil, nil, "sixth"},
	}
	assert.Equal(t, expected, result)
}

func TestSet_ComplexNestedWithArrays(t *testing.T) {
	set, err := NewSetFromPairs([]string{"org.teams[0].members[1].name=alice"})
	require.NoError(t, err)

	input := map[string]any{}
	result, err := set.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"org": map[string]any{
			"teams": []any{
				map[string]any{
					"members": []any{
						nil,
						map[string]any{"name": "alice"},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestSet_WhitespaceHandling(t *testing.T) {
	set, err := NewSetFromPairs([]string{"  name  =  alice  "})
	require.NoError(t, err)

	result, err := set.Apply(map[string]any{})
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestSet_JSONWithSpaces(t *testing.T) {
	set, err := NewSetFromPairs([]string{`data={ "key": "value" }`})
	require.NoError(t, err)

	result, err := set.Apply(map[string]any{})
	require.NoError(t, err)

	expected := map[string]any{
		"data": map[string]any{"key": "value"},
	}
	assert.Equal(t, expected, result)
}

func TestSet_SpecialCharactersInString(t *testing.T) {
	set, err := NewSetFromPairs([]string{`message=hello@world.com!`})
	require.NoError(t, err)

	result, err := set.Apply(map[string]any{})
	require.NoError(t, err)

	expected := map[string]any{"message": "hello@world.com!"}
	assert.Equal(t, expected, result)
}

func TestParseJSONish_Number(t *testing.T) {
	val, err := parseJSONish("42")
	require.NoError(t, err)
	assert.Equal(t, float64(42), val)
}

func TestParseJSONish_Float(t *testing.T) {
	val, err := parseJSONish("3.14")
	require.NoError(t, err)
	assert.Equal(t, 3.14, val)
}

func TestParseJSONish_BoolTrue(t *testing.T) {
	val, err := parseJSONish("true")
	require.NoError(t, err)
	assert.Equal(t, true, val)
}

func TestParseJSONish_BoolFalse(t *testing.T) {
	val, err := parseJSONish("false")
	require.NoError(t, err)
	assert.Equal(t, false, val)
}

func TestParseJSONish_Null(t *testing.T) {
	val, err := parseJSONish("null")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestParseJSONish_JSONString(t *testing.T) {
	val, err := parseJSONish(`"hello"`)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestParseJSONish_PlainString(t *testing.T) {
	val, err := parseJSONish("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestParseJSONish_JSONArray(t *testing.T) {
	val, err := parseJSONish(`[1,2,3]`)
	require.NoError(t, err)
	assert.Equal(t, []any{float64(1), float64(2), float64(3)}, val)
}

func TestParseJSONish_JSONObject(t *testing.T) {
	val, err := parseJSONish(`{"a":1}`)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"a": float64(1)}, val)
}

func TestSplitOnce_Found(t *testing.T) {
	left, right, ok := splitOnce("key=value", '=')
	assert.True(t, ok)
	assert.Equal(t, "key", left)
	assert.Equal(t, "value", right)
}

func TestSplitOnce_NotFound(t *testing.T) {
	left, right, ok := splitOnce("noequals", '=')
	assert.False(t, ok)
	assert.Equal(t, "noequals", left)
	assert.Equal(t, "", right)
}

func TestSplitOnce_Multiple(t *testing.T) {
	left, right, ok := splitOnce("a=b=c", '=')
	assert.True(t, ok)
	assert.Equal(t, "a", left)
	assert.Equal(t, "b=c", right)
}
