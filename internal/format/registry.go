package format

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

// Global registry of available formats
var (
	registry   = make(map[string]Format)
	registryMu sync.RWMutex
)

// Register adds a format to the global registry.
// This should typically be called from format package init() functions.
// If a format with the same name already exists, it will be replaced.
func Register(f Format) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[f.Name()] = f
}

// Get retrieves a format by name from the registry.
// Returns an error if the format is not registered.
func Get(name string) (Format, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	f, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("unknown format: %s", name)
	}
	return f, nil
}

// List returns the names of all registered formats.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// AutoDetect analyzes input data and returns the best matching format.
// It peeks at the first 1KB of data and runs all registered detectors.
// Returns the format with the highest confidence score.
// The returned bufio.Reader can be used to read the full input (peek is preserved).
func AutoDetect(r io.Reader) (Format, *bufio.Reader, error) {
	// Create buffered reader with generous peek buffer
	br := bufio.NewReaderSize(r, 64*1024)

	// Peek at beginning of input for format detection
	peek, err := br.Peek(1024)
	if err != nil && err != io.EOF {
		return nil, br, fmt.Errorf("failed to peek input: %w", err)
	}

	// If we got EOF on peek, the file is smaller than 1KB - that's ok
	// We'll just detect based on whatever we got

	registryMu.RLock()
	defer registryMu.RUnlock()

	var bestFormat Format
	var bestScore int

	// Run all detectors and find the one with highest confidence
	for _, f := range registry {
		score, err := f.Detector().Detect(peek)
		if err != nil {
			// Detection error - skip this format but don't fail completely
			continue
		}
		if score > bestScore {
			bestScore = score
			bestFormat = f
		}
	}

	if bestFormat == nil {
		return nil, br, fmt.Errorf("unable to detect format from input")
	}

	return bestFormat, br, nil
}
