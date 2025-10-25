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

// segment represents one step in a path. Either a key (map) and optional index (array).
// Examples:
//   - "user"                -> {key: "user", idx: nil}
//   - "items[0]"            -> {key: "items", idx: 0}
//   - "payload" or "data[1]": similar
type segment struct {
	key string
	idx *int // optional array index
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
		// We do not implement wildcards here.
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

		n, err := strconv.Atoi(idxStr)

		if err != nil || n < 0 {
			return nil, fmt.Errorf("invalid non-negative index in %q", part)
		}

		s.idx = &n
		segs = append(segs, s)
	}

	return segs, nil
}
