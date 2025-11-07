package format

import (
	"fmt"
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
