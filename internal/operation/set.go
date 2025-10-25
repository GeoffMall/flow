package operation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ----------------------------- Set -----------------------------

// Set holds multiple "path = value" assignments.
// Example pairs:
//   - "a.b=42"
//   - "user.name=\"alice\""   (or just user.name=alice → becomes string "alice")
//   - "flags.debug=true"
//   - "spec.image={\"name\":\"app\",\"tag\":\"v1\"}"
//   - "items[0]={\"id\":1}"
type Set struct {
	Assignments []Assignment
}

type Assignment struct {
	Path  string
	Value any
}

func NewSetFromPairs(pairs []string) (*Set, error) {
	as := make([]Assignment, 0, len(pairs))

	for _, p := range pairs {
		path, raw, ok := splitOnce(p, '=')
		if !ok {
			return nil, fmt.Errorf("invalid --set %q (expected path=value)", p)
		}

		path = strings.TrimSpace(path)

		if path == "" {
			return nil, fmt.Errorf("invalid --set %q: empty path", p)
		}

		val, err := parseJSONish(strings.TrimSpace(raw))

		if err != nil {
			return nil, fmt.Errorf("invalid --set %q: %w", p, err)
		}

		as = append(as, Assignment{Path: path, Value: val})
	}

	return &Set{Assignments: as}, nil
}

func (s *Set) Description() string {
	parts := make([]string, 0, len(s.Assignments))
	for _, a := range s.Assignments {
		parts = append(parts, a.Path)
	}

	return "set(" + strings.Join(parts, ", ") + ")"
}

func (s *Set) Apply(v any) (any, error) {
	// If the root isn't an object and we need to set a nested key, we convert to object.
	// If it's nil, start a new object.
	root, ok := v.(map[string]any)
	if !ok {
		// If caller pipes in a scalar/array and sets "": reject; we don't support empty/root path.
		// Otherwise, we start from an empty object and place values there.
		root = make(map[string]any)
	}

	for _, a := range s.Assignments {
		segs, err := parsePath(a.Path)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", a.Path, err)
		}

		setAtPathOverwrite(root, segs, a.Value)
	}

	return root, nil
}

// setAtPathOverwrite walks/creates maps/slices to place val at segs.
// If it encounters an incompatible existing type, it OVERWRITES it with the needed container.
func setAtPathOverwrite(root map[string]any, segs []segment, val any) {
	cur := root

	for i, s := range segs {
		isLast := i == len(segs)-1

		if s.idx == nil {
			cur = handleMapSegmentOverwrite(cur, s.key, val, isLast)
			if isLast {
				return
			}
		} else {
			cur = handleSliceSegmentOverwrite(cur, s.key, *s.idx, val, isLast)
			if isLast {
				return
			}
		}
	}
}

// handleMapSegmentOverwrite processes a segment without an index (simple map traversal)
func handleMapSegmentOverwrite(cur map[string]any, key string, val any, isLast bool) map[string]any {
	if isLast {
		cur[key] = val
		return cur
	}

	return ensureNestedMap(cur, key)
}

// handleSliceSegmentOverwrite processes a segment with an index (slice operation)
func handleSliceSegmentOverwrite(cur map[string]any, key string, idx int, val any, isLast bool) map[string]any {
	slice := ensureSliceOverwrite(cur, key)
	slice = expandSliceToIndex(slice, idx)

	if isLast {
		slice[idx] = val
		cur[key] = slice
		return cur
	}

	nextMap := ensureMapAtSliceIndex(slice, idx)
	cur[key] = slice
	return nextMap
}

// ensureNestedMap gets or creates a map under the given key, overwriting if necessary
func ensureNestedMap(cur map[string]any, key string) map[string]any {
	child, exists := cur[key]
	if next, ok := child.(map[string]any); exists && ok && next != nil {
		return next
	}

	next := make(map[string]any)
	cur[key] = next
	return next
}

// ensureSliceOverwrite gets or creates a slice under the given key, overwriting if necessary
func ensureSliceOverwrite(cur map[string]any, key string) []any {
	child, exists := cur[key]
	if slice, ok := child.([]any); exists && ok && slice != nil {
		return slice
	}

	return make([]any, 0)
}

// expandSliceToIndex grows the slice to accommodate the given index
func expandSliceToIndex(slice []any, idx int) []any {
	if idx < len(slice) {
		return slice
	}

	newSlice := make([]any, idx+1)
	copy(newSlice, slice)
	return newSlice
}

// ensureMapAtSliceIndex gets or creates a map at the specified slice index
func ensureMapAtSliceIndex(slice []any, idx int) map[string]any {
	if next, ok := slice[idx].(map[string]any); ok && next != nil {
		return next
	}

	next := make(map[string]any)
	slice[idx] = next
	return next
}

// parseJSONish tries to unmarshal JSON first (so numbers/bools/objects/arrays work).
// If it fails, the raw string is returned as a plain string.
func parseJSONish(s string) (any, error) {
	// If it's clearly not JSON (no quotes/brackets/braces, not a number/bool/null),
	// we can return it as-is for speed. But correctness is more important; try JSON.
	var v any
	if err := json.Unmarshal([]byte(s), &v); err == nil {
		return v, nil
	}

	// Not valid JSON → treat as plain string. That's convenient for --set k=foo
	// (no need to write k="foo").
	return s, nil
}

// splitOnce splits on the first sep. Returns (left, right, true) if sep found.
func splitOnce(s string, sep byte) (string, string, bool) {
	i := strings.IndexByte(s, sep)
	if i < 0 {
		return s, "", false
	}

	return s[:i], s[i+1:], true
}
