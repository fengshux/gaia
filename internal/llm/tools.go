// Package llm provides LLM provider implementations
package llm

import (
	"partner/pkg/types"
)

// ToolRegistry manages tool registrations
type ToolRegistry struct {
	tools     map[string]types.ToolDefinition
	executors map[string]types.ToolExecutor
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:     make(map[string]types.ToolDefinition),
		executors: make(map[string]types.ToolExecutor),
	}
}

// Register registers a tool with its executor
func (r *ToolRegistry) Register(tool types.ToolDefinition, executor types.ToolExecutor) {
	r.tools[tool.Name] = tool
	r.executors[tool.Name] = executor
}

// GetTool returns a tool by name
func (r *ToolRegistry) GetTool(name string) (types.ToolDefinition, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetExecutor returns an executor by tool name
func (r *ToolRegistry) GetExecutor(name string) (types.ToolExecutor, bool) {
	executor, ok := r.executors[name]
	return executor, ok
}

// GetAllTools returns all registered tools
func (r *ToolRegistry) GetAllTools() []types.ToolDefinition {
	tools := make([]types.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute executes a tool by name
func (r *ToolRegistry) Execute(name string, params map[string]interface{}) (interface{}, error) {
	executor, ok := r.GetExecutor(name)
	if !ok {
		return nil, &ToolNotFoundError{Name: name}
	}
	return executor.Execute(nil, params)
}

// ToolNotFoundError is returned when a tool is not found
type ToolNotFoundError struct {
	Name string
}

func (e *ToolNotFoundError) Error() string {
	return "tool not found: " + e.Name
}
