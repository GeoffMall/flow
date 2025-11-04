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
	Paths             []string
	PreserveHierarchy bool // if true, preserves full path structure (legacy behavior)
}

func NewPick(paths []string, preserveHierarchy bool) *Pick {
	return &Pick{
		Paths:             paths,
		PreserveHierarchy: preserveHierarchy,
	}
}

func (p *Pick) Description() string {
	return "pick(" + strings.Join(p.Paths, ", ") + ")"
}

// Apply returns a document based on the requested paths and mode.
// In jq-like mode (default): returns raw values or flattened objects.
// In preserve-hierarchy mode: returns full path structure (legacy behavior).
func (p *Pick) Apply(v any) (any, error) {
	// If no paths requested, return input as-is.
	if len(p.Paths) == 0 {
		return v, nil
	}

	// Legacy behavior: preserve full hierarchy
	if p.PreserveHierarchy {
		return p.applyWithHierarchy(v)
	}

	// New jq-like behavior
	if len(p.Paths) == 1 {
		return p.applySinglePath(v, p.Paths[0])
	}
	return p.applyMultiplePaths(v)
}

// applySinglePath extracts a single path and returns just the value (or array for wildcards).
// This matches jq behavior: jq '.result[0].domain' returns just "pure-skin.name"
func (p *Pick) applySinglePath(v any, pathStr string) (any, error) {
	// Check if path contains wildcard
	hasWildcard := strings.Contains(pathStr, "[*]")

	// Expand wildcards
	expandedPaths, err := expandWildcardPaths(v, pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid --pick %q: %w", pathStr, err)
	}

	// Wildcard that expanded to 0 items - return empty array
	if hasWildcard && len(expandedPaths) == 0 {
		return []any{}, nil
	}

	// No wildcards or single concrete path
	if len(expandedPaths) <= 1 {
		segs, err := parsePath(pathStr)
		if err != nil {
			return nil, err
		}
		val, ok := getAtPath(v, segs)
		if !ok {
			return nil, nil // Return null for missing paths (jq behavior)
		}
		return val, nil // JUST THE VALUE
	}

	// Wildcard produced multiple values - return array
	var results []any
	for _, expandedPath := range expandedPaths {
		segs, err := parsePath(expandedPath)
		if err != nil {
			return nil, err
		}
		val, ok := getAtPath(v, segs)
		if ok {
			results = append(results, val)
		}
	}
	return results, nil
}

// applyMultiplePaths extracts multiple paths and returns a flattened object.
// Example: --pick user.name --pick user.age returns {"name": "alice", "age": 30}
func (p *Pick) applyMultiplePaths(v any) (any, error) {
	out := make(map[string]any)

	for _, pathStr := range p.Paths {
		expandedPaths, err := expandWildcardPaths(v, pathStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --pick %q: %w", pathStr, err)
		}

		if len(expandedPaths) == 0 {
			expandedPaths = []string{pathStr}
		}

		// Wildcard case: collect into array
		if len(expandedPaths) > 1 {
			if err := p.addWildcardResults(v, pathStr, expandedPaths, out); err != nil {
				return nil, err
			}
			continue
		}

		// Single path case
		if err := p.addSinglePathResult(v, pathStr, out); err != nil {
			return nil, err
		}
	}

	if len(out) == 0 {
		return nil, nil // Return null if nothing found
	}

	return out, nil
}

// addWildcardResults collects wildcard expansion results into an array
func (p *Pick) addWildcardResults(v any, pathStr string, expandedPaths []string, out map[string]any) error {
	var results []any
	for _, expandedPath := range expandedPaths {
		segs, err := parsePath(expandedPath)
		if err != nil {
			return err
		}
		val, ok := getAtPath(v, segs)
		if ok {
			results = append(results, val)
		}
	}
	if len(results) > 0 {
		finalKey := getFinalKeyFromPath(pathStr)
		out[finalKey] = results
	}
	return nil
}

// addSinglePathResult adds a single path's value to the output with flattened key
func (p *Pick) addSinglePathResult(v any, pathStr string, out map[string]any) error {
	segs, err := parsePath(pathStr)
	if err != nil {
		return err
	}

	val, ok := getAtPath(v, segs)
	if !ok {
		return nil // Skip missing paths in multi-pick
	}

	// Extract just the final key name for flattening
	finalKey := getFinalKey(segs)
	out[finalKey] = val
	return nil
}

// applyWithHierarchy implements the legacy behavior that preserves full path structure.
func (p *Pick) applyWithHierarchy(v any) (any, error) {
	out := make(map[string]any)

	for _, raw := range p.Paths {
		// Expand wildcards into concrete paths
		expandedPaths, err := expandWildcardPaths(v, raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --pick %q: %w", raw, err)
		}

		// If no expansion occurred (no wildcards), expandedPaths will contain the original path
		if len(expandedPaths) == 0 {
			expandedPaths = []string{raw}
		}

		// Process each expanded path
		for _, expandedPath := range expandedPaths {
			segs, err := parsePath(expandedPath)
			if err != nil {
				return nil, fmt.Errorf("invalid expanded path %q: %w", expandedPath, err)
			}

			val, ok := getAtPath(v, segs)
			if !ok {
				continue
			}

			// Merge into output at the same path (preserving hierarchy)
			setAtPath(out, segs, val)
		}
	}

	// If nothing was picked, return empty object.
	return out, nil
}

// getFinalKey extracts the last key name from a parsed path (for flattening).
// Example: segments for "user.name" -> returns "name"
func getFinalKey(segs []segment) string {
	if len(segs) == 0 {
		return ""
	}
	lastSeg := segs[len(segs)-1]
	return lastSeg.key
}

// getFinalKeyFromPath extracts the last key name from a path string.
// Handles wildcards: "items[*].name" -> "name"
func getFinalKeyFromPath(pathStr string) string {
	// Remove wildcard notation
	pathStr = strings.ReplaceAll(pathStr, "[*]", "")
	// Get last segment after last dot
	parts := strings.Split(pathStr, ".")
	return parts[len(parts)-1]
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
