package operation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Operation represents a transformation applied to a document.
// If you already declared this interface elsewhere, remove this block.
type Operation interface {
	// Apply takes an input document v (decoded JSON/YAML as Go values)
	// and returns a transformed document.
	Apply(v any) (any, error)
	Description() string
}

// ----------------------------- Path parsing -----------------------------

// A segment represents one step in a path. Either a key (map) and optional index (array).
// Examples:
//   - "user"                -> {key: "user", idx: nil}
//   - "items[0]"            -> {key: "items", idx: 0}
//   - "items[*]"            -> {key: "items", idx: -1} (wildcard)
type segment struct {
	key string
	idx *int // optional array index, -1 for wildcard
}

func parsePath(path string) ([]segment, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	parts := strings.Split(path, ".")
	segs := make([]segment, 0, len(parts))

	for _, part := range parts {
		// Allowed shapes:
		//   - key
		//   - key[idx]
		//   - key[*] (wildcard)
		s := segment{}

		// Look for bracketed index
		open := strings.IndexByte(part, '[')
		if open < 0 {
			// Just a key
			s.key = part
			segs = append(segs, s)
			continue
		}

		// key[idx] expected
		if !strings.HasSuffix(part, "]") || open == 0 {
			return nil, fmt.Errorf("invalid segment %q", part)
		}

		s.key = part[:open]
		idxStr := part[open+1 : len(part)-1]

		if idxStr == "" {
			return nil, fmt.Errorf("empty index in %q", part)
		}

		// Handle wildcard
		if idxStr == "*" {
			wildcardIdx := -1
			s.idx = &wildcardIdx
		} else {
			n, err := strconv.Atoi(idxStr)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid non-negative index in %q", part)
			}
			s.idx = &n
		}
		segs = append(segs, s)
	}

	return segs, nil
}

// ----------------------------- Wildcard expansion -----------------------------

// expandWildcardPaths takes a path with wildcards and returns all concrete paths
func expandWildcardPaths(v any, pathStr string) ([]string, error) {
	segs, err := parsePath(pathStr)
	if err != nil {
		return nil, err
	}

	return expandSegments(v, segs, "")
}

// expandSegments recursively expands wildcard segments into concrete paths
func expandSegments(v any, segs []segment, currentPath string) ([]string, error) {
	if len(segs) == 0 {
		return []string{currentPath}, nil
	}

	seg := segs[0]
	remaining := segs[1:]

	// Handle map key
	if seg.key != "" {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, nil // Path doesn't exist
		}

		child, exists := m[seg.key]
		if !exists {
			return nil, nil
		}

		newPath := buildPath(currentPath, seg.key)

		// If no index, continue with child
		if seg.idx == nil {
			return expandSegments(child, remaining, newPath)
		}

		// Handle array index (potentially wildcard)
		return expandArrayIndex(child, seg.idx, remaining, newPath)
	}

	return nil, nil
}

// expandArrayIndex handles array indexing with wildcard support
func expandArrayIndex(v any, idx *int, remaining []segment, currentPath string) ([]string, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	// Check if this is a wildcard
	if *idx == -1 {
		var allPaths []string
		for i := 0; i < len(arr); i++ {
			indexPath := fmt.Sprintf("%s[%d]", currentPath, i)
			expandedPaths, err := expandSegments(arr[i], remaining, indexPath)
			if err != nil {
				return nil, err
			}
			allPaths = append(allPaths, expandedPaths...)
		}
		return allPaths, nil
	}

	// Regular index
	if *idx < 0 || *idx >= len(arr) {
		return nil, nil
	}

	indexPath := fmt.Sprintf("%s[%d]", currentPath, *idx)
	return expandSegments(arr[*idx], remaining, indexPath)
}

func buildPath(current, key string) string {
	if current == "" {
		return key
	}
	return current + "." + key
}
