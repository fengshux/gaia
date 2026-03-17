// Package plugin provides plugin management
package plugin

import (
	"context"
)

// Plugin defines the interface for plugins
type Plugin interface {
	// Info returns plugin information
	Info() PluginInfo

	// Tools returns the tools provided by this plugin
	Tools() []Tool

	// Init initializes the plugin with configuration
	Init(ctx context.Context, config map[string]interface{}) error

	// Close cleans up plugin resources
	Close() error
}

// PluginInfo contains plugin metadata
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

// Tool defines a tool that can be executed
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Parameters returns the JSON schema for parameters
	Parameters() interface{}

	// Execute runs the tool with given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// BaseTool provides a base implementation for tools
type BaseTool struct {
	name        string
	description string
	parameters  interface{}
	executor    func(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// NewBaseTool creates a new base tool
func NewBaseTool(name, description string, parameters interface{}, executor func(ctx context.Context, params map[string]interface{}) (interface{}, error)) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
		executor:    executor,
	}
}

// Name returns the tool name
func (t *BaseTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *BaseTool) Description() string {
	return t.description
}

// Parameters returns the JSON schema for parameters
func (t *BaseTool) Parameters() interface{} {
	return t.parameters
}

// Execute runs the tool
func (t *BaseTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return t.executor(ctx, params)
}
