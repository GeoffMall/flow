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

// ----------------------------- Setting logic -----------------------------

// setAtPathOverwrite walks/creates maps/slices to place val at segs.
// If it encounters an incompatible existing type, it OVERWRITES it with the needed container.
func setAtPathOverwrite(root map[string]any, segs []segment, val any) {
	cur := root
	for i, s := range segs {
		isLast := i == len(segs)-1

		// Ensure a child slot for s.key
		child, exists := cur[s.key]

		// If no index: it's a straight object chain
		if s.idx == nil {
			if isLast {
				cur[s.key] = val
				return
			}
			// Child must be a map; overwrite if not
			next, ok := child.(map[string]any)
			if !exists || !ok || next == nil {
				next = make(map[string]any)
				cur[s.key] = next
			}
			cur = next
			continue
		}

		// key[index] case: child must be a slice; overwrite if not
		sl, ok := child.([]any)
		if !exists || !ok || sl == nil {
			sl = make([]any, 0)
		}

		// grow slice to index+1
		if *s.idx >= len(sl) {
			newSl := make([]any, *s.idx+1)
			copy(newSl, sl)
			sl = newSl
		}

		if isLast {
			sl[*s.idx] = val
			cur[s.key] = sl
			return
		}

		// Need a map at this index for further nesting
		next, ok := sl[*s.idx].(map[string]any)
		if !ok || next == nil {
			next = make(map[string]any)
			sl[*s.idx] = next
		}

		cur[s.key] = sl
		cur = next
	}
}

// ----------------------------- Value parsing -----------------------------

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
