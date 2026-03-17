// Package types provides common types used across the assistant
package types

import (
	"context"
	"fmt"
	"time"
)

// Role represents the role of a message sender
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// Message represents a chat message
type Message struct {
	ID        string                 `json:"id"`
	Role      Role                   `json:"role"`
	Content   string                 `json:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	ToolID    string                 `json:"tool_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// ToolCall represents a tool call request
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string      `json:"tool_call_id"`
	Content    string      `json:"content"`
	IsError    bool        `json:"is_error"`
	Data       interface{} `json:"data,omitempty"`
}

// Session represents a conversation session
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AddMessage adds a message to the session
func (s *Session) AddMessage(msg Message) {
	if msg.ID == "" {
		msg.ID = generateID()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
}

// generateID generates a simple ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ToolDefinition defines a tool that can be used by the LLM
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"` // JSON Schema
}

// ToolExecutor is the interface for executing tools
type ToolExecutor interface {
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// PermissionLevel represents the permission level for an action
type PermissionLevel string

const (
	PermissionAsk   PermissionLevel = "ask"
	PermissionAuto  PermissionLevel = "auto"
	PermissionDeny  PermissionLevel = "deny"
)

// PermissionRequest represents a request for permission
type PermissionRequest struct {
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Params      map[string]interface{} `json:"params,omitempty"`
	Risk        string                 `json:"risk"` // low, medium, high
	Description string                 `json:"description"`
}

// PermissionResponse represents the response to a permission request
type PermissionResponse struct {
	Granted bool   `json:"granted"`
	Reason  string `json:"reason,omitempty"`
}

// Response represents a response from the assistant
type Response struct {
	Content    string       `json:"content"`
	ToolCalls  []ToolCall   `json:"tool_calls,omitempty"`
	ToolResult *ToolResult  `json:"tool_result,omitempty"`
	Done       bool         `json:"done"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content   string     `json:"content,omitempty"`
	ToolCall  *ToolCall  `json:"tool_call,omitempty"`
	Done      bool       `json:"done"`
	Error     error      `json:"error,omitempty"`
}
