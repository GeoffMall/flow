package operation

import (
	"fmt"
	"strings"
)

// filteredMarker is a sentinel value returned by WHERE when a document doesn't match.
// This allows us to distinguish between "filtered out" and "actual nil/null value".
type filteredMarker struct{}

// Filtered is the singleton instance of filteredMarker used to indicate filtered documents.
var Filtered = &filteredMarker{}

// Where filters documents based on key=value conditions.
// If a document doesn't match all conditions, it returns Filtered (which filters it out).
// All conditions are AND'ed together.
type Where struct {
	conditions []whereCondition
}

// whereCondition represents a single key=value filter condition.
type whereCondition struct {
	path  []segment
	value string
}

// NewWhere creates a new Where operation from a list of key=value pairs.
// Example: NewWhere([]string{"user.name=Alice", "status=active"})
func NewWhere(pairs []string) (*Where, error) {
	if len(pairs) == 0 {
		return &Where{conditions: nil}, nil
	}

	conditions := make([]whereCondition, 0, len(pairs))

	for _, pair := range pairs {
		// Parse key=value
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid where condition '%s': must be in format key=value", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("invalid where condition '%s': key cannot be empty", pair)
		}

		// Parse the path
		path, err := parsePath(key)
		if err != nil {
			return nil, fmt.Errorf("invalid where condition '%s': %w", pair, err)
		}

		conditions = append(conditions, whereCondition{
			path:  path,
			value: value,
		})
	}

	return &Where{conditions: conditions}, nil
}

// Apply filters the document based on all WHERE conditions.
// Returns Filtered if the document doesn't match (filters it out).
// Returns the original document if it matches all conditions.
func (w *Where) Apply(v any) (any, error) {
	// If no conditions, pass through
	if len(w.conditions) == 0 {
		return v, nil
	}

	// Check each condition
	for _, condition := range w.conditions {
		// Navigate to the field
		fieldValue, err := navigatePath(v, condition.path)
		if err != nil {
			// Field doesn't exist or path invalid - doesn't match
			return Filtered, nil
		}

		// Convert field value to string for comparison
		fieldStr := fmt.Sprintf("%v", fieldValue)

		// Check if it matches the expected value
		if fieldStr != condition.value {
			// Doesn't match - filter out
			return Filtered, nil
		}
	}

	// All conditions matched - keep the document
	return v, nil
}

// Description returns a human-readable description of this operation.
func (w *Where) Description() string {
	if len(w.conditions) == 0 {
		return "where: (no conditions)"
	}

	parts := make([]string, 0, len(w.conditions))
	for _, cond := range w.conditions {
		pathStr := pathToString(cond.path)
		parts = append(parts, fmt.Sprintf("%s=%s", pathStr, cond.value))
	}

	return fmt.Sprintf("where: %s", strings.Join(parts, " AND "))
}

// navigatePath walks through the document following the path segments.
// Returns the value at the end of the path, or an error if the path is invalid.
//
//nolint:cyclop // Path navigation requires checking multiple type cases
func navigatePath(v any, path []segment) (any, error) {
	current := v

	for _, seg := range path {
		switch c := current.(type) {
		case map[string]any:
			val, ok := c[seg.key]
			if !ok {
				return nil, fmt.Errorf("field '%s' not found", seg.key)
			}
			current = val

			// Handle array index if specified
			if seg.idx != nil {
				arr, ok := current.([]any)
				if !ok {
					return nil, fmt.Errorf("field '%s' is not an array", seg.key)
				}
				idx := *seg.idx
				if idx < 0 || idx >= len(arr) {
					return nil, fmt.Errorf("index %d out of range for array '%s'", idx, seg.key)
				}
				current = arr[idx]
			}

		case []any:
			// If current is an array, we need an index
			if seg.idx == nil {
				return nil, fmt.Errorf("array requires index")
			}
			idx := *seg.idx
			if idx < 0 || idx >= len(c) {
				return nil, fmt.Errorf("index %d out of range", idx)
			}
			current = c[idx]

		default:
			return nil, fmt.Errorf("cannot navigate through %T", current)
		}
	}

	return current, nil
}

// pathToString converts a path back to string representation for display.
func pathToString(path []segment) string {
	var parts []string
	for _, seg := range path {
		if seg.idx != nil {
			parts = append(parts, fmt.Sprintf("%s[%d]", seg.key, *seg.idx))
		} else {
			parts = append(parts, seg.key)
		}
	}
	return strings.Join(parts, ".")
}
