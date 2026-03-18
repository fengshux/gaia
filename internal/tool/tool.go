// Package tool provides built-in tools
package tool

import (
	"context"

	"gaia/pkg/types"
)

// Tool interface for tools
type Tool interface {
	Name() string
	Description() string
	Parameters() interface{}
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// BaseTool provides a base implementation
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

// ToDefinition converts a Tool to ToolDefinition
func ToDefinition(t Tool) types.ToolDefinition {
	return types.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters:  t.Parameters(),
	}
}

// Common JSON Schema definitions

// StringParam defines a string parameter
type StringParam struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// NumberParam defines a number parameter
type NumberParam struct {
	Type        string  `json:"type"`
	Description string  `json:"description,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
}

// BooleanParam defines a boolean parameter
type BooleanParam struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ArrayParam defines an array parameter
type ArrayParam struct {
	Type        string      `json:"type"`
	Items       interface{} `json:"items"`
	Description string      `json:"description,omitempty"`
}

// ObjectParam defines an object parameter
type ObjectParam struct {
	Type        string                 `json:"type"`
	Properties  map[string]interface{} `json:"properties"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}
