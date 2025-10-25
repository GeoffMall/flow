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

// ----------------------------- Set value -----------------------------

// setAtPath mutates 'out' by creating/intersecting maps/slices to place 'val' at 'segs'.
func setAtPath(out map[string]any, segs []segment, val any) {
	curObj := out

	for i, s := range segs {
		isLast := i == len(segs)-1

		// ensure map for key
		child, exists := curObj[s.key]
		if s.idx == nil {
			// simple object chain: map -> map -> ... -> value
			if isLast {
				curObj[s.key] = val
				return
			}
			// ensure next map
			m, ok := child.(map[string]any)
			if !exists || !ok {
				m = make(map[string]any)
				curObj[s.key] = m
			}
			curObj = m
			continue
		}

		// key with index: ensure slice under the key
		slice, ok := child.([]any)
		if !exists || !ok {
			slice = make([]any, 0)
		}

		// grow slice to idx+1
		if *s.idx >= len(slice) {
			newSlice := make([]any, *s.idx+1)
			copy(newSlice, slice)
			slice = newSlice
		}

		if isLast {
			// set value directly at index
			slice[*s.idx] = val
			curObj[s.key] = slice
			return
		}

		// need a map at that index for further nested steps
		nextObj, ok := slice[*s.idx].(map[string]any)
		if !ok || nextObj == nil {
			nextObj = make(map[string]any)
			slice[*s.idx] = nextObj
		}

		curObj[s.key] = slice
		curObj = nextObj
	}
}
