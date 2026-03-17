// Package plugin provides plugin management
package plugin

import (
	"sync"
)

// Registry provides a global plugin registry
type Registry struct {
	mu      sync.RWMutex
	factories map[string]Factory
}

// Factory is a function that creates a plugin
type Factory func() Plugin

// globalRegistry is the global plugin registry
var globalRegistry = &Registry{
	factories: make(map[string]Factory),
}

// RegisterFactory registers a plugin factory
func RegisterFactory(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// GetFactory returns a plugin factory by name
func GetFactory(name string) (Factory, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	f, ok := globalRegistry.factories[name]
	return f, ok
}

// ListFactories returns all registered factory names
func ListFactories() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.factories))
	for name := range globalRegistry.factories {
		names = append(names, name)
	}
	return names
}

// CreatePlugin creates a plugin by name
func CreatePlugin(name string) (Plugin, bool) {
	factory, ok := GetFactory(name)
	if !ok {
		return nil, false
	}
	return factory(), true
}
