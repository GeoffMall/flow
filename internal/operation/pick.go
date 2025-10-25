package operation

import (
	"fmt"
	"strings"
)

// Pick holds a list of dotted paths to extract from the input document.
// Examples:
//   - "user.name"
//   - "items[0].name"
//   - "a.b[12].c"
type Pick struct {
	Paths []string
}

func NewPick(paths []string) *Pick { return &Pick{Paths: paths} }

func (p *Pick) Description() string {
	return "pick(" + strings.Join(p.Paths, ", ") + ")"
}

// Apply returns a new document containing only the requested paths (merged).
// If a path doesn't exist in the source, it's ignored.
func (p *Pick) Apply(v any) (any, error) {
	// If no paths requested, return input as-is.
	if len(p.Paths) == 0 {
		return v, nil
	}

	// The result is generally an object; if a path targets the root (""), we’ll just return v.
	out := make(map[string]any)

	for _, raw := range p.Paths {
		segs, err := parsePath(raw)
		if err != nil {
			// Surface path parsing errors; caller may want to show friendly message.
			return nil, fmt.Errorf("invalid --pick %q: %w", raw, err)
		}

		val, ok := getAtPath(v, segs)
		if !ok {
			// Path doesn't exist; skip
			continue
		}

		// Merge into output at the same path.
		setAtPath(out, segs, val)
	}

	// If nothing was picked, return empty object.
	return out, nil
}

// ----------------------------- Get value -----------------------------

// getAtPath walks the input structure according to segs and returns (value, true)
// if found, otherwise (nil, false). It never mutates the input.
func getAtPath(v any, segs []segment) (any, bool) {
	cur := v

	for _, s := range segs {
		// Step 1: key on maps
		if s.key != "" {
			m, ok := asStringMap(cur)
			if !ok {
				return nil, false
			}

			next, ok := m[s.key]

			if !ok {
				return nil, false
			}

			cur = next
		}

		// Step 2: optional array index
		if s.idx != nil {
			arr, ok := asSlice(cur)
			if !ok {
				return nil, false
			}

			if *s.idx < 0 || *s.idx >= len(arr) {
				return nil, false
			}

			cur = arr[*s.idx]
		}
	}

	return cur, true
}

func asStringMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	// Some decoders (YAML) may yield map[interface{}]interface{} before normalization;
	// by design in this project, parser.normalizeYAML already fixed that.
	return nil, false
}

func asSlice(v any) ([]any, bool) {
	if a, ok := v.([]any); ok {
		return a, true
	}

	return nil, false
}

// setAtPath mutates 'out' by creating/intersecting maps/slices to place 'val' at 'segs'.
func setAtPath(out map[string]any, segs []segment, val any) {
	curObj := out

	for i, s := range segs {
		isLast := i == len(segs)-1

		if s.idx == nil {
			curObj = handleMapSegment(curObj, s.key, val, isLast)
			if isLast {
				return
			}
		} else {
			curObj = handleSliceSegment(curObj, s.key, *s.idx, val, isLast)
			if isLast {
				return
			}
		}
	}
}

// handleMapSegment processes a segment without an index (simple map traversal)
func handleMapSegment(curObj map[string]any, key string, val any, isLast bool) map[string]any {
	if isLast {
		curObj[key] = val
		return curObj
	}

	// Ensure next map exists
	child, exists := curObj[key]
	m, ok := child.(map[string]any)
	if !exists || !ok {
		m = make(map[string]any)
		curObj[key] = m
	}

	return m
}

// handleSliceSegment processes a segment with an index (slice operation)
func handleSliceSegment(curObj map[string]any, key string, idx int, val any, isLast bool) map[string]any {
	slice := ensureSliceExists(curObj, key)
	slice = growSliceIfNeeded(slice, idx)

	if isLast {
		slice[idx] = val
		curObj[key] = slice
		return curObj
	}

	// Need a map at that index for further nested steps
	nextObj := ensureMapAtIndex(slice, idx)
	curObj[key] = slice
	return nextObj
}

// ensureSliceExists gets or creates a slice under the given key
func ensureSliceExists(curObj map[string]any, key string) []any {
	child, exists := curObj[key]
	slice, ok := child.([]any)
	if !exists || !ok {
		slice = make([]any, 0)
	}
	return slice
}

// growSliceIfNeeded expands the slice to accommodate the given index
func growSliceIfNeeded(slice []any, idx int) []any {
	if idx >= len(slice) {
		newSlice := make([]any, idx+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	return slice
}

// ensureMapAtIndex gets or creates a map at the specified slice index
func ensureMapAtIndex(slice []any, idx int) map[string]any {
	nextObj, ok := slice[idx].(map[string]any)
	if !ok || nextObj == nil {
		nextObj = make(map[string]any)
		slice[idx] = nextObj
	}
	return nextObj
}
