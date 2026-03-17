// Package mcp implements the Model Context Protocol
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// ToolHandler handles tool execution
type ToolHandler func(ctx context.Context, params map[string]interface{}) (*ToolCallResult, error)

// Server represents an MCP server
type Server struct {
	mu          sync.RWMutex
	info        ServerInfo
	tools       map[string]ToolDefinition
	toolHandlers map[string]ToolHandler
	resources   map[string]Resource
	prompts     map[string]Prompt
	clients     map[string]*websocket.Conn
	upgrader    websocket.Upgrader
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		info: ServerInfo{
			Name:    name,
			Version: version,
		},
		tools:       make(map[string]ToolDefinition),
		toolHandlers: make(map[string]ToolHandler),
		resources:   make(map[string]Resource),
		prompts:     make(map[string]Prompt),
		clients:     make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// RegisterTool registers a tool with its handler
func (s *Server) RegisterTool(tool ToolDefinition, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.toolHandlers[tool.Name] = handler
}

// UnregisterTool removes a tool
func (s *Server) UnregisterTool(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tools, name)
	delete(s.toolHandlers, name)
}

// RegisterResource registers a resource
func (s *Server) RegisterResource(resource Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[resource.URI] = resource
}

// RegisterPrompt registers a prompt
func (s *Server) RegisterPrompt(prompt Prompt) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prompts[prompt.Name] = prompt
}

// HandleRequest handles a JSON-RPC request
func (s *Server) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	default:
		return NewErrorResponse(req.ID, MethodNotFound, "method not found: "+req.Method, nil), nil
	}
}

// HandleNotification handles a JSON-RPC notification
func (s *Server) HandleNotification(ctx context.Context, notif *Notification) error {
	// Handle notifications if needed
	return nil
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *Request) (*Response, error) {
	var params InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, InvalidParams, "invalid params", nil), nil
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo:      s.info,
		Capabilities: Capabilities{
			Tools: &ToolCapabilities{Supported: true},
			Resources: &ResourceCapabilities{Supported: true},
			Prompts: &PromptCapabilities{Supported: true},
		},
	}

	return NewResponse(req.ID, result), nil
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *Request) (*Response, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]ToolDefinition, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return NewResponse(req.ID, map[string]interface{}{
		"tools": tools,
	}), nil
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(ctx context.Context, req *Request) (*Response, error) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, InvalidParams, "invalid params", nil), nil
	}

	s.mu.RLock()
	handler, ok := s.toolHandlers[params.Name]
	s.mu.RUnlock()

	if !ok {
		return NewErrorResponse(req.ID, InvalidParams, "tool not found: "+params.Name, nil), nil
	}

	result, err := handler(ctx, params.Arguments)
	if err != nil {
		return NewResponse(req.ID, &ToolCallResult{
			Content: []Content{
				{Type: "text", Text: err.Error()},
			},
			IsError: true,
		}), nil
	}

	return NewResponse(req.ID, result), nil
}

// handleResourcesList handles the resources/list request
func (s *Server) handleResourcesList(req *Request) (*Response, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}

	return NewResponse(req.ID, map[string]interface{}{
		"resources": resources,
	}), nil
}

// handlePromptsList handles the prompts/list request
func (s *Server) handlePromptsList(req *Request) (*Response, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}

	return NewResponse(req.ID, map[string]interface{}{
		"prompts": prompts,
	}), nil
}

// Notify sends a notification to all connected clients
func (s *Server) Notify(method string, params interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notif, err := NewNotification(method, params)
	if err != nil {
		return err
	}

	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	for _, conn := range s.clients {
		conn.WriteMessage(websocket.TextMessage, data)
	}

	return nil
}

// StartHTTP starts the server as an HTTP server with WebSocket support
func (s *Server) StartHTTP(addr string) error {
	http.HandleFunc("/mcp", s.handleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return http.ListenAndServe(addr, nil)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// Register client
	clientID := fmt.Sprintf("%p", conn)
	s.mu.Lock()
	s.clients[clientID] = conn
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
	}()

	// Create transport and process messages
	transport := NewWebSocketTransport(conn)
	ProcessTransportMessages(r.Context(), transport, s)
}
