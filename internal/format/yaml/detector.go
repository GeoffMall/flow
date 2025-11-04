package yaml

import (
	"strings"
)

// Detector implements format.Detector for YAML format.
type Detector struct{}

// Detect analyzes input bytes to determine if they contain YAML data.
// Returns a confidence score from 0-100:
//   - 100: Definitely YAML (starts with ---, %YAML, or has clear key: value)
//   - 0: Not YAML (looks like JSON)
func (d *Detector) Detect(peek []byte) (int, error) {
	head := strings.TrimLeft(string(peek), " \t\r\n")

	// Empty input - could be either, but default to JSON (return low score)
	if len(head) == 0 {
		return 0, nil
	}

	// Definite YAML indicators
	if head[0] == '%' {
		// YAML directive like "%YAML 1.2"
		return 100, nil
	}

	if strings.HasPrefix(head, "---") {
		// YAML document separator
		return 100, nil
	}

	// Definite JSON indicators (not YAML)
	firstChar := head[0]
	if firstChar == '{' || firstChar == '[' {
		return 0, nil
	}

	// Heuristic check: YAML typically has ':' before ',' or '}'
	if looksLikeYAML(head) {
		return 90, nil
	}

	// Default: not YAML
	return 0, nil
}

// looksLikeYAML checks if the content has YAML-like key: value structure.
// This is a heuristic: if the first line contains ':' before ',' or '}',
// it's likely YAML key: value format.
func looksLikeYAML(head string) bool {
	// Get first line
	line := head
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}

	// Look for colon (key: value pattern)
	colon := strings.IndexByte(line, ':')
	if colon < 0 {
		return false
	}

	// Find comma and closing brace
	comma := strings.IndexByte(line, ',')
	closeBrace := strings.IndexByte(line, '}')

	// If comma/brace not found, use large value
	if comma == -1 {
		comma = 1 << 30
	}
	if closeBrace == -1 {
		closeBrace = 1 << 30
	}

	// YAML: colon appears before JSON structural characters
	return colon < comma && colon < closeBrace
}
