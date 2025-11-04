package operation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPipeline(t *testing.T) {
	pick := NewPick([]string{"name"}, true)
	set, _ := NewSetFromPairs([]string{"age=30"})

	pipe := NewPipeline(pick, set)
	assert.NotNil(t, pipe)
	assert.Len(t, pipe.Ops, 2)
}

func TestPipeline_Empty(t *testing.T) {
	pipe := NewPipeline()
	assert.True(t, pipe.Empty())
}

func TestPipeline_NotEmpty(t *testing.T) {
	pick := NewPick([]string{"name"}, true)
	pipe := NewPipeline(pick)
	assert.False(t, pipe.Empty())
}

func TestPipeline_Append(t *testing.T) {
	pipe := NewPipeline()
	assert.True(t, pipe.Empty())

	pick := NewPick([]string{"name"}, true)
	pipe.Append(pick)
	assert.False(t, pipe.Empty())
	assert.Len(t, pipe.Ops, 1)

	set, _ := NewSetFromPairs([]string{"age=30"})
	pipe.Append(set)
	assert.Len(t, pipe.Ops, 2)
}

func TestPipeline_ApplyEmpty(t *testing.T) {
	pipe := NewPipeline()
	input := map[string]any{"name": "alice"}
	result, err := pipe.Apply(input)
	require.NoError(t, err)
	assert.Equal(t, input, result)
}

func TestPipeline_ApplySingleOp(t *testing.T) {
	pick := NewPick([]string{"name"}, true)
	pipe := NewPipeline(pick)

	input := map[string]any{"name": "alice", "age": 30}
	result, err := pipe.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestPipeline_ApplyMultipleOps(t *testing.T) {
	pick := NewPick([]string{"name", "age"}, true)
	set, _ := NewSetFromPairs([]string{"city=NYC"})
	pipe := NewPipeline(pick, set)

	input := map[string]any{
		"name":  "alice",
		"age":   30,
		"email": "alice@example.com",
	}
	result, err := pipe.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"name": "alice",
		"age":  30,
		"city": "NYC",
	}
	assert.Equal(t, expected, result)
}

func TestPipeline_ApplyChaining(t *testing.T) {
	// Pick name and email, then delete email, then set age
	pick := NewPick([]string{"name", "email"}, true)
	del := NewDelete([]string{"email"})
	set, _ := NewSetFromPairs([]string{"age=30"})
	pipe := NewPipeline(pick, del, set)

	input := map[string]any{
		"name":  "alice",
		"email": "alice@example.com",
		"city":  "NYC",
	}
	result, err := pipe.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"name": "alice",
		"age":  float64(30), // Set operation parses "30" as JSON, which becomes float64
	}
	assert.Equal(t, expected, result)
}

func TestPipeline_ErrorInFirstOp(t *testing.T) {
	pick := NewPick([]string{""}, false) // Invalid path
	pipe := NewPipeline(pick)

	input := map[string]any{"name": "alice"}
	_, err := pipe.Apply(input)
	require.Error(t, err)

	var stepErr StepError
	assert.True(t, errors.As(err, &stepErr))
	assert.Equal(t, 0, stepErr.Index)
	assert.Contains(t, stepErr.OpDesc, "pick")
}

func TestPipeline_ErrorInMiddleOp(t *testing.T) {
	pick := NewPick([]string{"name"}, true)
	set, _ := NewSetFromPairs([]string{"age="}) // Valid but creates empty string
	invalidPick := NewPick([]string{""}, false) // Invalid path
	pipe := NewPipeline(pick, set, invalidPick)

	input := map[string]any{"name": "alice"}
	_, err := pipe.Apply(input)
	require.Error(t, err)

	var stepErr StepError
	assert.True(t, errors.As(err, &stepErr))
	assert.Equal(t, 2, stepErr.Index)
}

func TestPipeline_ErrorUnwrap(t *testing.T) {
	pick := NewPick([]string{""}, false)
	pipe := NewPipeline(pick)

	input := map[string]any{"name": "alice"}
	_, err := pipe.Apply(input)
	require.Error(t, err)

	var stepErr StepError
	require.True(t, errors.As(err, &stepErr))

	unwrapped := stepErr.Unwrap()
	assert.NotNil(t, unwrapped)
	assert.Contains(t, unwrapped.Error(), "empty path")
}

func TestStepError_ErrorMessage_WithDesc(t *testing.T) {
	innerErr := errors.New("something went wrong")
	stepErr := StepError{
		Index:   2,
		OpDesc:  "pick(name)",
		Wrapped: innerErr,
	}

	msg := stepErr.Error()
	assert.Contains(t, msg, "pipeline step 2")
	assert.Contains(t, msg, "pick(name)")
	assert.Contains(t, msg, "something went wrong")
}

func TestStepError_ErrorMessage_WithoutDesc(t *testing.T) {
	innerErr := errors.New("something went wrong")
	stepErr := StepError{
		Index:   1,
		OpDesc:  "",
		Wrapped: innerErr,
	}

	msg := stepErr.Error()
	assert.Contains(t, msg, "pipeline step 1")
	assert.Contains(t, msg, "something went wrong")
	assert.NotContains(t, msg, "()")
}

// Mock operation for testing
type mockOp struct {
	desc      string
	transform func(any) (any, error)
}

func (m *mockOp) Apply(v any) (any, error) {
	if m.transform != nil {
		return m.transform(v)
	}
	return v, nil
}

func (m *mockOp) Description() string {
	return m.desc
}

func TestPipeline_CustomOperation(t *testing.T) {
	// Test with a mock operation
	mock := &mockOp{
		desc: "mock(test)",
		transform: func(v any) (any, error) {
			if m, ok := v.(map[string]any); ok {
				m["custom"] = true
				return m, nil
			}
			return v, nil
		},
	}

	pipe := NewPipeline(mock)
	input := map[string]any{"name": "alice"}
	result, err := pipe.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice", "custom": true}
	assert.Equal(t, expected, result)
}

func TestPipeline_OperationReturnsError(t *testing.T) {
	mock := &mockOp{
		desc: "failing-op",
		transform: func(v any) (any, error) {
			return nil, errors.New("intentional error")
		},
	}

	pipe := NewPipeline(mock)
	input := map[string]any{"name": "alice"}
	_, err := pipe.Apply(input)
	require.Error(t, err)

	var stepErr StepError
	require.True(t, errors.As(err, &stepErr))
	assert.Equal(t, "failing-op", stepErr.OpDesc)
	assert.Contains(t, stepErr.Wrapped.Error(), "intentional error")
}

func TestPipeline_MultipleOperationsWithTransformation(t *testing.T) {
	// Create a pipeline that transforms data through multiple steps
	addField := &mockOp{
		desc: "add-field",
		transform: func(v any) (any, error) {
			if m, ok := v.(map[string]any); ok {
				m["step1"] = true
				return m, nil
			}
			return v, nil
		},
	}

	modifyField := &mockOp{
		desc: "modify-field",
		transform: func(v any) (any, error) {
			if m, ok := v.(map[string]any); ok {
				m["step2"] = "modified"
				return m, nil
			}
			return v, nil
		},
	}

	pipe := NewPipeline(addField, modifyField)
	input := map[string]any{"original": "data"}
	result, err := pipe.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{
		"original": "data",
		"step1":    true,
		"step2":    "modified",
	}
	assert.Equal(t, expected, result)
}

// Test safeDesc function indirectly
type panickyOp struct{}

func (p *panickyOp) Apply(v any) (any, error) {
	return v, nil
}

func (p *panickyOp) Description() string {
	panic("description panics!")
}

func TestPipeline_SafeDesc_HandlesPanic(t *testing.T) {
	// safeDesc should handle panics in Description()
	panicky := &panickyOp{}
	pipe := NewPipeline(panicky)

	// Force an error to trigger StepError creation with safeDesc
	mock := &mockOp{
		desc: "will-fail",
		transform: func(v any) (any, error) {
			return nil, errors.New("error")
		},
	}
	pipe.Append(mock)

	input := map[string]any{"test": "data"}
	_, err := pipe.Apply(input)
	require.Error(t, err)

	// Should not panic even though Description() panics
	var stepErr StepError
	assert.True(t, errors.As(err, &stepErr))
}

func TestPipeline_DefensiveCopy(t *testing.T) {
	// Verify that NewPipeline makes a defensive copy
	pick := NewPick([]string{"name"}, true)
	set, _ := NewSetFromPairs([]string{"age=30"})

	ops := []Operation{pick, set}
	pipe := NewPipeline(ops...)

	// Modify the original slice
	ops[0] = nil

	// Pipeline should still work with its copy
	input := map[string]any{"name": "alice", "email": "alice@example.com"}
	result, err := pipe.Apply(input)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice", "age": float64(30)}
	assert.Equal(t, expected, result)
}

func TestPipeline_NilInput(t *testing.T) {
	set, _ := NewSetFromPairs([]string{"name=alice"})
	pipe := NewPipeline(set)

	result, err := pipe.Apply(nil)
	require.NoError(t, err)

	expected := map[string]any{"name": "alice"}
	assert.Equal(t, expected, result)
}

func TestPipeline_ComplexRealWorldScenario(t *testing.T) {
	// Simulate a real-world pipeline:
	// 1. Pick specific fields from input
	// 2. Delete sensitive fields
	// 3. Add metadata
	pick := NewPick([]string{"user.name", "user.email", "user.age"}, true)
	del := NewDelete([]string{"user.password"}) // This won't exist after pick, but testing the flow
	set, _ := NewSetFromPairs([]string{
		"metadata.processed=true",
		"metadata.version=1",
	})

	pipe := NewPipeline(pick, del, set)

	input := map[string]any{
		"user": map[string]any{
			"name":     "alice",
			"email":    "alice@example.com",
			"age":      30,
			"password": "secret123",
		},
		"system": map[string]any{
			"id": 12345,
		},
	}

	result, err := pipe.Apply(input)
	require.NoError(t, err)

	// Should have picked user fields (excluding password), and added metadata
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)

	assert.Contains(t, resultMap, "user")
	assert.Contains(t, resultMap, "metadata")

	userMap := resultMap["user"].(map[string]any)
	assert.Equal(t, "alice", userMap["name"])
	assert.Equal(t, "alice@example.com", userMap["email"])
	assert.Equal(t, 30, userMap["age"])
	assert.NotContains(t, userMap, "password")

	metadataMap := resultMap["metadata"].(map[string]any)
	assert.Equal(t, true, metadataMap["processed"])
	assert.Equal(t, float64(1), metadataMap["version"])
}
