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
		// Expand wildcards into concrete paths
		expandedPaths, err := expandWildcardPaths(v, raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --delete %q: %w", raw, err)
		}

		// If no expansion occurred, use original path
		if len(expandedPaths) == 0 {
			expandedPaths = []string{raw}
		}

		// Process each expanded path
		for _, expandedPath := range expandedPaths {
			segs, err := parsePath(expandedPath)
			if err != nil {
				return nil, fmt.Errorf("invalid expanded path %q: %w", expandedPath, err)
			}

			deleteAtPath(&v, segs)
		}
	}

	return v, nil
}

// deleteAtPath walks 'v' by segs and deletes the targeted node if present.
// It mutates 'v' in place when it's a map or slice. If path doesn't exist, no-op.
func deleteAtPath(v *any, segs []segment) {
	if len(segs) == 0 || v == nil {
		return
	}

	parent, err := navigateToParent(v, segs[:len(segs)-1])
	if err != nil {
		return
	}

	deleteFromParent(parent, segs[len(segs)-1])
}

// navigateToParent walks through all segments except the last one
func navigateToParent(v *any, segs []segment) (*any, error) {
	cur := v

	for _, s := range segs {
		next, err := stepIntoMap(cur, s.key)
		if err != nil {
			return nil, err
		}

		if s.idx != nil {
			next, err = stepIntoSlice(next, *s.idx)
			if err != nil {
				return nil, err
			}
		}

		*cur = next
		cur = &next
	}

	return cur, nil
}

// stepIntoMap steps into a map by key
func stepIntoMap(cur *any, key string) (any, error) {
	m, ok := (*cur).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("not a map")
	}

	child, exists := m[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	return child, nil
}

// stepIntoSlice steps into a slice by index
func stepIntoSlice(val any, idx int) (any, error) {
	arr, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("not a slice")
	}

	if idx < 0 || idx >= len(arr) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return arr[idx], nil
}

// deleteFromParent deletes the final segment from its parent container
func deleteFromParent(parent *any, last segment) {
	m, ok := (*parent).(map[string]any)
	if !ok {
		return
	}

	if last.idx == nil {
		delete(m, last.key)
		return
	}

	deleteFromSliceInMap(m, last.key, *last.idx)
}

// deleteFromSliceInMap deletes an element from a slice stored in a map
func deleteFromSliceInMap(m map[string]any, key string, idx int) {
	child, exists := m[key]
	if !exists {
		return
	}

	arr, ok := child.([]any)
	if !ok {
		return
	}

	if idx < 0 || idx >= len(arr) {
		return
	}

	// Remove arr[idx] by shifting left
	arr = append(arr[:idx], arr[idx+1:]...)
	m[key] = arr
}
