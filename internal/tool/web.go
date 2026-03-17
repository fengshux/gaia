// Package tool provides built-in tools
package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebFetchTool fetches content from a URL
type WebFetchTool struct {
	client *http.Client
}

// NewWebFetchTool creates a new web fetch tool
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *WebFetchTool) Name() string {
	return "web_fetch"
}

func (t *WebFetchTool) Description() string {
	return "Fetch content from a URL"
}

func (t *WebFetchTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"url": StringParam{
				Type:        "string",
				Description: "The URL to fetch",
			},
			"method": StringParam{
				Type:        "string",
				Description: "HTTP method (default: GET)",
				Enum:        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			},
			"headers": ObjectParam{
				Type:        "object",
				Description: "HTTP headers as key-value pairs",
				Properties:  map[string]interface{}{},
			},
			"body": StringParam{
				Type:        "string",
				Description: "Request body (for POST/PUT/PATCH)",
			},
			"timeout": NumberParam{
				Type:        "number",
				Description: "Timeout in seconds (default: 30)",
			},
		},
		Required: []string{"url"},
	}
}

func (t *WebFetchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	rawURL, ok := params["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url parameter is required")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure scheme
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	method := "GET"
	if m, ok := params["method"].(string); ok {
		method = strings.ToUpper(m)
	}

	// Create request
	var body io.Reader
	if b, ok := params["body"].(string); ok {
		body = strings.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Set timeout
	timeout := 30 * time.Second
	if t, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(t) * time.Second
	}

	client := &http.Client{Timeout: timeout}

	// Execute request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    resp.Header,
		"body":       string(respBody),
		"duration":   time.Since(start).String(),
	}, nil
}

// WebSearchTool searches the web (placeholder - requires API key)
type WebSearchTool struct{}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

func (t *WebSearchTool) Name() string {
	return "web_search"
}

func (t *WebSearchTool) Description() string {
	return "Search the web for information"
}

func (t *WebSearchTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"query": StringParam{
				Type:        "string",
				Description: "The search query",
			},
			"num_results": NumberParam{
				Type:        "number",
				Description: "Number of results to return (default: 5)",
			},
		},
		Required: []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required")
	}

	_ = 5 // numResults placeholder - would be used with real API

	// This is a placeholder - in production, you'd use a real search API
	// like Google Custom Search, Bing, or DuckDuckGo
	return map[string]interface{}{
		"query":   query,
		"results": []interface{}{},
		"message": "Web search requires API configuration",
	}, nil
}

// WebAPITool makes API calls
type WebAPITool struct{}

// NewWebAPITool creates a new web API tool
func NewWebAPITool() *WebAPITool {
	return &WebAPITool{}
}

func (t *WebAPITool) Name() string {
	return "web_api"
}

func (t *WebAPITool) Description() string {
	return "Make API calls with JSON support"
}

func (t *WebAPITool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"url": StringParam{
				Type:        "string",
				Description: "The API endpoint URL",
			},
			"method": StringParam{
				Type:        "string",
				Description: "HTTP method",
				Enum:        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			},
			"headers": ObjectParam{
				Type:        "object",
				Description: "HTTP headers",
				Properties:  map[string]interface{}{},
			},
			"params": ObjectParam{
				Type:        "object",
				Description: "Query parameters",
				Properties:  map[string]interface{}{},
			},
			"data": ObjectParam{
				Type:        "object",
				Description: "JSON request body",
				Properties:  map[string]interface{}{},
			},
		},
		Required: []string{"url", "method"},
	}
}

func (t *WebAPITool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	rawURL, ok := params["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url parameter is required")
	}

	method, ok := params["method"].(string)
	if !ok {
		method = "GET"
	}

	// Build URL with query params
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if queryParams, ok := params["params"].(map[string]interface{}); ok {
		q := parsedURL.Query()
		for k, v := range queryParams {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		parsedURL.RawQuery = q.Encode()
	}

	// Create request body
	var body io.Reader
	if data, ok := params["data"].(map[string]interface{}); ok {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add headers
	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Execute
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response
	var result interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &result)

	return map[string]interface{}{
		"status": resp.StatusCode,
		"data":   result,
		"raw":    string(respBody),
	}, nil
}
