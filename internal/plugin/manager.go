// Package plugin provides plugin management
package plugin

import (
	"context"
	"fmt"
	"sync"

	"partner/pkg/types"
)

// Manager manages plugins and their tools
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	tools   map[string]Tool
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		tools:   make(map[string]Tool),
	}
}

// Register registers a plugin
func (m *Manager) Register(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := p.Info()
	if _, exists := m.plugins[info.Name]; exists {
		return fmt.Errorf("plugin %s already registered", info.Name)
	}

	m.plugins[info.Name] = p

	// Register tools
	for _, tool := range p.Tools() {
		m.tools[tool.Name()] = tool
	}

	return nil
}

// Unregister removes a plugin
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Remove tools
	for _, tool := range p.Tools() {
		delete(m.tools, tool.Name())
	}

	// Close plugin
	if err := p.Close(); err != nil {
		return fmt.Errorf("failed to close plugin: %w", err)
	}

	delete(m.plugins, name)
	return nil
}

// GetPlugin returns a plugin by name
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.plugins[name]
	return p, ok
}

// GetTool returns a tool by name
func (m *Manager) GetTool(name string) (Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tools[name]
	return t, ok
}

// GetAllTools returns all registered tools
func (m *Manager) GetAllTools() []types.ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]types.ToolDefinition, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, types.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		})
	}
	return tools
}

// ExecuteTool executes a tool by name
func (m *Manager) ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	t, ok := m.tools[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return t.Execute(ctx, params)
}

// InitPlugin initializes a plugin with configuration
func (m *Manager) InitPlugin(ctx context.Context, name string, config map[string]interface{}) error {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	return p.Init(ctx, config)
}

// Close closes all plugins
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, p := range m.plugins {
		if err := p.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close plugin %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing plugins: %v", errs)
	}

	return nil
}

// ListPlugins returns all registered plugins
func (m *Manager) ListPlugins() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(m.plugins))
	for _, p := range m.plugins {
		infos = append(infos, p.Info())
	}
	return infos
}
