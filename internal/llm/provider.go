// Package llm provides LLM provider implementations
package llm

import (
	"context"

	"assistant/pkg/types"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// StreamChat sends a chat request and returns a stream of responses
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan types.StreamChunk, error)

	// GetTools returns the available tools
	GetTools() []types.ToolDefinition

	// RegisterTool registers a tool
	RegisterTool(tool types.ToolDefinition, executor types.ToolExecutor)

	// Close closes the provider
	Close() error
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Model       string          `json:"model"`
	Messages    []types.Message `json:"messages"`
	Tools       []types.ToolDefinition `json:"tools,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	ID      string          `json:"id"`
	Model   string          `json:"model"`
	Message types.Message   `json:"message"`
	Usage   Usage           `json:"usage"`
	Done    bool            `json:"done"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	Temperature float64
	MaxTokens   int
}
