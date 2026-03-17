// Package mcp implements the Model Context Protocol
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// Client represents an MCP client
type Client struct {
	mu            sync.RWMutex
	info          ClientInfo
	transport     Transport
	requestID     atomic.Uint64
	pending       map[uint64]chan *Response
	tools         map[string]ToolDefinition
	resources     map[string]Resource
	prompts       map[string]Prompt
	connected     bool
}

// NewClient creates a new MCP client
func NewClient(name, version string) *Client {
	return &Client{
		info: ClientInfo{
			Name:    name,
			Version: version,
		},
		pending:   make(map[uint64]chan *Response),
		tools:     make(map[string]ToolDefinition),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
	}
}

// Connect connects to an MCP server using the given transport
func (c *Client) Connect(ctx context.Context, transport Transport) error {
	c.mu.Lock()
	c.transport = transport
	c.mu.Unlock()

	// Start message processing
	go c.processMessages(ctx)

	// Initialize connection
	result, err := c.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	// Store capabilities
	_ = result
	c.connected = true

	return nil
}

// Initialize sends an initialize request
func (c *Client) Initialize(ctx context.Context) (*InitializeResult, error) {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		ClientInfo:      c.info,
		Capabilities: Capabilities{
			Tools:     &ToolCapabilities{Supported: true},
			Resources: &ResourceCapabilities{Supported: true},
			Prompts:   &PromptCapabilities{Supported: true},
		},
	}

	resp, err := c.sendRequest(ctx, "initialize", params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	var result InitializeResult
	data, _ := json.Marshal(resp.Result)
	json.Unmarshal(data, &result)

	return &result, nil
}

// ListTools requests the list of available tools
func (c *Client) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	resp, err := c.sendRequest(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}

	var result struct {
		Tools []ToolDefinition `json:"tools"`
	}
	data, _ := json.Marshal(resp.Result)
	json.Unmarshal(data, &result)

	// Store tools
	c.mu.Lock()
	for _, tool := range result.Tools {
		c.tools[tool.Name] = tool
	}
	c.mu.Unlock()

	return result.Tools, nil
}

// CallTool calls a tool on the server
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolCallResult, error) {
	params := ToolCallParams{
		Name:      name,
		Arguments: args,
	}

	resp, err := c.sendRequest(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/call error: %s", resp.Error.Message)
	}

	var result ToolCallResult
	data, _ := json.Marshal(resp.Result)
	json.Unmarshal(data, &result)

	return &result, nil
}

// ListResources requests the list of available resources
func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	resp, err := c.sendRequest(ctx, "resources/list", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("resources/list error: %s", resp.Error.Message)
	}

	var result struct {
		Resources []Resource `json:"resources"`
	}
	data, _ := json.Marshal(resp.Result)
	json.Unmarshal(data, &result)

	// Store resources
	c.mu.Lock()
	for _, resource := range result.Resources {
		c.resources[resource.URI] = resource
	}
	c.mu.Unlock()

	return result.Resources, nil
}

// ListPrompts requests the list of available prompts
func (c *Client) ListPrompts(ctx context.Context) ([]Prompt, error) {
	resp, err := c.sendRequest(ctx, "prompts/list", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("prompts/list error: %s", resp.Error.Message)
	}

	var result struct {
		Prompts []Prompt `json:"prompts"`
	}
	data, _ := json.Marshal(resp.Result)
	json.Unmarshal(data, &result)

	// Store prompts
	c.mu.Lock()
	for _, prompt := range result.Prompts {
		c.prompts[prompt.Name] = prompt
	}
	c.mu.Unlock()

	return result.Prompts, nil
}

// GetTool returns a tool by name
func (c *Client) GetTool(name string) (ToolDefinition, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tool, ok := c.tools[name]
	return tool, ok
}

// GetTools returns all available tools
func (c *Client) GetTools() []ToolDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tools := make([]ToolDefinition, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Close closes the client connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// sendRequest sends a request and waits for response
func (c *Client) sendRequest(ctx context.Context, method string, params interface{}) (*Response, error) {
	id := c.requestID.Add(1)

	req, err := NewRequest(id, method, params)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Create response channel
	respCh := make(chan *Response, 1)
	c.mu.Lock()
	c.pending[id] = respCh
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	// Send request
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil, fmt.Errorf("not connected")
	}

	if err := transport.Send(data); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// processMessages processes incoming messages
func (c *Client) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.mu.RLock()
			transport := c.transport
			c.mu.RUnlock()

			if transport == nil {
				return
			}

			data, err := transport.Receive()
			if err != nil {
				fmt.Printf("receive error: %v\n", err)
				return
			}

			// Try to parse as response
			var resp Response
			if err := json.Unmarshal(data, &resp); err == nil && resp.ID != nil {
				// It's a response
				var id uint64
				switch v := resp.ID.(type) {
				case float64:
					id = uint64(v)
				case uint64:
					id = v
				case int:
					id = uint64(v)
				}

				c.mu.Lock()
				if ch, ok := c.pending[id]; ok {
					ch <- &resp
				}
				c.mu.Unlock()
			} else {
				// It's a notification
				var notif Notification
				if err := json.Unmarshal(data, &notif); err == nil {
					c.handleNotification(ctx, &notif)
				}
			}
		}
	}
}

// handleNotification handles incoming notifications
func (c *Client) handleNotification(ctx context.Context, notif *Notification) {
	// Handle notifications if needed
	switch notif.Method {
	case "tools/list_changed":
		// Refresh tools list
		c.ListTools(ctx)
	case "resources/list_changed":
		// Refresh resources list
		c.ListResources(ctx)
	}
}
