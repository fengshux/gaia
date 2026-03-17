// Package plugin provides plugin management
package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"assistant/pkg/utils"
)

// Loader loads plugins from filesystem
type Loader struct {
	pluginDir string
}

// NewLoader creates a new plugin loader
func NewLoader(pluginDir string) *Loader {
	return &Loader{
		pluginDir: pluginDir,
	}
}

// LoadAll loads all plugins from the plugin directory
func (l *Loader) LoadAll(ctx context.Context, manager *Manager) error {
	// Ensure plugin directory exists
	if !utils.DirExists(l.pluginDir) {
		return fmt.Errorf("plugin directory does not exist: %s", l.pluginDir)
	}

	// Walk through plugin directory
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only load .so files
		if filepath.Ext(entry.Name()) != ".so" {
			continue
		}

		// Load plugin
		pluginPath := filepath.Join(l.pluginDir, entry.Name())
		if err := l.Load(ctx, manager, pluginPath); err != nil {
			// Log error but continue loading other plugins
			fmt.Printf("Warning: failed to load plugin %s: %v\n", pluginPath, err)
		}
	}

	return nil
}

// Load loads a single plugin from file
func (l *Loader) Load(ctx context.Context, manager *Manager, path string) error {
	// Open plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for Plugin symbol
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin does not export Plugin symbol: %w", err)
	}

	// Type assert to Plugin interface
	pluginInstance, ok := sym.(Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement Plugin interface")
	}

	// Register plugin
	if err := manager.Register(pluginInstance); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	// Initialize plugin
	if err := pluginInstance.Init(ctx, nil); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	return nil
}

// LoadBuiltin loads built-in plugins
func (l *Loader) LoadBuiltin(ctx context.Context, manager *Manager, enabled []string) error {
	for _, name := range enabled {
		// Check if factory exists
		factory, ok := GetFactory(name)
		if !ok {
			fmt.Printf("Warning: built-in plugin %s not found\n", name)
			continue
		}

		// Create plugin instance
		p := factory()

		// Register plugin
		if err := manager.Register(p); err != nil {
			fmt.Printf("Warning: failed to register plugin %s: %v\n", name, err)
			continue
		}

		// Initialize plugin
		if err := p.Init(ctx, nil); err != nil {
			fmt.Printf("Warning: failed to initialize plugin %s: %v\n", name, err)
		}
	}

	return nil
}
