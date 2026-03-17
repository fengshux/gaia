// Package core provides the core engine
package core

import (
	"context"
	"fmt"
	"sync"

	"assistant/internal/config"
	"assistant/internal/llm"
	"assistant/internal/plugin"
	"assistant/pkg/types"
)

// Engine is the main AI assistant engine
type Engine struct {
	mu       sync.RWMutex
	config   *config.Config
	provider llm.Provider
	plugins  *plugin.Manager
	sessions *SessionManager
	perm     *PermissionManager
	running  bool
}

// NewEngine creates a new engine
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		config:   cfg,
		plugins:  plugin.NewManager(),
		sessions: NewSessionManager(),
		perm:     NewPermissionManager(cfg.Permissions),
	}
}

// Initialize initializes the engine
func (e *Engine) Initialize(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize LLM provider
	e.provider = llm.NewOpenAIProvider(llm.ProviderConfig{
		APIKey:      e.config.LLM.APIKey,
		BaseURL:     e.config.LLM.BaseURL,
		Model:       e.config.LLM.Model,
		Temperature: e.config.LLM.Temperature,
		MaxTokens:   e.config.LLM.MaxTokens,
	})

	// Load built-in plugins
	loader := plugin.NewLoader(e.config.Plugins.Directory)
	if err := loader.LoadBuiltin(ctx, e.plugins, e.config.Plugins.Enabled); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Register tools with LLM provider
	for _, tool := range e.plugins.GetAllTools() {
		toolCopy := tool
		e.provider.RegisterTool(toolCopy, &toolExecutorAdapter{
			manager: e.plugins,
			name:    toolCopy.Name,
		})
	}

	e.running = true
	return nil
}

// Process processes a user input
func (e *Engine) Process(ctx context.Context, input string) (*types.Response, error) {
	return e.ProcessWithSession(ctx, input, "")
}

// ProcessWithSession processes input with a specific session
func (e *Engine) ProcessWithSession(ctx context.Context, input, sessionID string) (*types.Response, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.running {
		return nil, fmt.Errorf("engine not initialized")
	}

	// Get or create session
	session := e.sessions.GetOrCreate(sessionID)

	// Add user message
	session.AddMessage(types.Message{
		Role:    types.RoleUser,
		Content: input,
	})

	// Build request
	req := &llm.ChatRequest{
		Model:    e.config.LLM.Model,
		Messages: session.Messages,
		Tools:    e.plugins.GetAllTools(),
	}

	// Call LLM
	resp, err := e.provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM error: %w", err)
	}

	// Handle tool calls
	if len(resp.Message.ToolCalls) > 0 {
		return e.handleToolCalls(ctx, session, resp.Message.ToolCalls)
	}

	// Add assistant message
	session.AddMessage(resp.Message)

	return &types.Response{
		Content: resp.Message.Content,
		Done:    true,
	}, nil
}

// ProcessStream processes input with streaming response
func (e *Engine) ProcessStream(ctx context.Context, input string) (<-chan types.StreamChunk, error) {
	return e.ProcessStreamWithSession(ctx, input, "")
}

// ProcessStreamWithSession processes input with streaming for a specific session
func (e *Engine) ProcessStreamWithSession(ctx context.Context, input, sessionID string) (<-chan types.StreamChunk, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.running {
		return nil, fmt.Errorf("engine not initialized")
	}

	// Get or create session
	session := e.sessions.GetOrCreate(sessionID)

	// Add user message
	session.AddMessage(types.Message{
		Role:    types.RoleUser,
		Content: input,
	})

	// Build request
	req := &llm.ChatRequest{
		Model:    e.config.LLM.Model,
		Messages: session.Messages,
		Tools:    e.plugins.GetAllTools(),
	}

	// Call LLM with streaming
	stream, err := e.provider.StreamChat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM error: %w", err)
	}

	// Create output channel
	out := make(chan types.StreamChunk, 100)

	// Process stream
	go func() {
		defer close(out)

		var content string
		var toolCalls []types.ToolCall

		for chunk := range stream {
			if chunk.Error != nil {
				out <- types.StreamChunk{Error: chunk.Error}
				return
			}

			if chunk.ToolCall != nil {
				toolCalls = append(toolCalls, *chunk.ToolCall)
			}

			content += chunk.Content
			out <- chunk

			if chunk.Done {
				break
			}
		}

		// Add assistant message
		session.AddMessage(types.Message{
			Role:      types.RoleAssistant,
			Content:   content,
			ToolCalls: toolCalls,
		})

		// Handle tool calls if any
		if len(toolCalls) > 0 {
			result, err := e.handleToolCalls(ctx, session, toolCalls)
			if err != nil {
				out <- types.StreamChunk{Error: err}
				return
			}

			// Continue processing if there's more content
			if result.Content != "" {
				out <- types.StreamChunk{Content: result.Content, Done: result.Done}
			}
		}
	}()

	return out, nil
}

// handleToolCalls handles tool calls from the LLM
func (e *Engine) handleToolCalls(ctx context.Context, session *types.Session, calls []types.ToolCall) (*types.Response, error) {
	var results []types.ToolResult

	for _, call := range calls {
		// Check permission
		if !e.perm.CheckPermission(call.Name, call.Arguments) {
			results = append(results, types.ToolResult{
				ToolCallID: call.ID,
				Content:    "Permission denied",
				IsError:    true,
			})
			continue
		}

		// Execute tool
		result, err := e.plugins.ExecuteTool(ctx, call.Name, call.Arguments)
		if err != nil {
			results = append(results, types.ToolResult{
				ToolCallID: call.ID,
				Content:    err.Error(),
				IsError:    true,
			})
			continue
		}

		// Convert result to string
		content := fmt.Sprintf("%v", result)
		results = append(results, types.ToolResult{
			ToolCallID: call.ID,
			Content:    content,
			Data:       result,
		})
	}

	// Add tool results to session
	for _, r := range results {
		session.AddMessage(types.Message{
			Role:    types.RoleTool,
			Content: r.Content,
			ToolID:  r.ToolCallID,
		})
	}

	// Continue conversation with LLM
	req := &llm.ChatRequest{
		Model:    e.config.LLM.Model,
		Messages: session.Messages,
		Tools:    e.plugins.GetAllTools(),
	}

	resp, err := e.provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM error: %w", err)
	}

	// Handle more tool calls recursively
	if len(resp.Message.ToolCalls) > 0 {
		return e.handleToolCalls(ctx, session, resp.Message.ToolCalls)
	}

	// Add final response
	session.AddMessage(resp.Message)

	return &types.Response{
		Content: resp.Message.Content,
		Done:    true,
	}, nil
}

// Close shuts down the engine
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.running = false

	if e.provider != nil {
		e.provider.Close()
	}

	return e.plugins.Close()
}

// GetSessionManager returns the session manager
func (e *Engine) GetSessionManager() *SessionManager {
	return e.sessions
}

// GetPluginManager returns the plugin manager
func (e *Engine) GetPluginManager() *plugin.Manager {
	return e.plugins
}

// toolExecutorAdapter adapts plugin manager to ToolExecutor interface
type toolExecutorAdapter struct {
	manager *plugin.Manager
	name    string
}

func (a *toolExecutorAdapter) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return a.manager.ExecuteTool(ctx, a.name, params)
}
