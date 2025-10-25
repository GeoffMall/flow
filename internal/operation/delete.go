package operation

import (
	"fmt"
	"strings"
)

// Delete holds a list of paths to remove from the input document.
// Examples:
//
//	--delete user.password
//	--delete items[0].meta.secret
//	--delete flags.debug
type Delete struct {
	Paths []string
}

func NewDelete(paths []string) *Delete { return &Delete{Paths: paths} }

func (d *Delete) Description() string {
	return "delete(" + strings.Join(d.Paths, ", ") + ")"
}

// Apply deletes each requested path from the input document in order.
// If the root value is not an object/array where a path begins, the
// corresponding delete is ignored (no-op). Returns the mutated value.
func (d *Delete) Apply(v any) (any, error) {
	// We mutate in place if the root is a map[string]any or []any.
	// If the root is scalar and paths target subfields, this becomes a no-op.
	for _, raw := range d.Paths {
		segs, err := parsePath(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --delete %q: %w", raw, err)
		}
		deleteAtPath(&v, segs)
	}
	return v, nil
}

// deleteAtPath walks 'v' by segs and deletes the targeted node if present.
// It mutates 'v' in place when it's a map or slice. If path doesn't exist, no-op.
func deleteAtPath(v *any, segs []segment) {
	if len(segs) == 0 || v == nil {
		return
	}
	cur := v

	// Walk until the parent of the final segment
	for i := 0; i < len(segs)-1; i++ {
		s := segs[i]

		// Step on map by key
		m, ok := (*cur).(map[string]any)
		if !ok {
			// Not a map → cannot proceed
			return
		}
		child, exists := m[s.key]
		if !exists {
			return
		}

		// Optional index into slice under that key
		if s.idx != nil {
			arr, ok := child.([]any)
			if !ok {
				return
			}
			if *s.idx < 0 || *s.idx >= len(arr) {
				return
			}
			// Descend into that element
			elem := arr[*s.idx]
			child = elem
		}

		// Update cur to reference the child slot so subsequent steps can mutate it
		*cur = child
	}

	// Now 'cur' is the parent container; delete the last segment from it.
	last := segs[len(segs)-1]

	// Parent must be a map for the last step's key
	m, ok := (*cur).(map[string]any)
	if !ok {
		return
	}
	// If last has no index: delete the key directly.
	if last.idx == nil {
		delete(m, last.key)
		return
	}

	// Otherwise last is key[index]: delete that element from the slice under the key.
	child, exists := m[last.key]
	if !exists {
		return
	}
	arr, ok := child.([]any)
	if !ok {
		return
	}
	idx := *last.idx
	if idx < 0 || idx >= len(arr) {
		return
	}
	// Remove arr[idx] by shifting left
	arr = append(arr[:idx], arr[idx+1:]...)
	m[last.key] = arr
}
