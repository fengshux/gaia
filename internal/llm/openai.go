// Package llm provides LLM provider implementations
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"partner/pkg/types"

	"github.com/go-resty/resty/v2"
)

// OpenAIProvider implements Provider for OpenAI-compatible APIs
type OpenAIProvider struct {
	client    *resty.Client
	config    ProviderConfig
	registry  *ToolRegistry
	converter *MessageConverter
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config ProviderConfig) *OpenAIProvider {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)
	client.SetAuthToken(config.APIKey)
	client.SetHeader("Content-Type", "application/json")

	return &OpenAIProvider{
		client:    client,
		config:    config,
		registry:  NewToolRegistry(),
		converter: &MessageConverter{},
	}
}

// Chat sends a chat request
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Build OpenAI request
	oaiReq := &openAIChatRequest{
		Model:       p.config.Model,
		Messages:    p.converter.ToOpenAIMessages(req.Messages),
		Temperature: p.config.Temperature,
		MaxTokens:   p.config.MaxTokens,
	}

	if len(req.Tools) > 0 {
		oaiReq.Tools = ToOpenAITools(req.Tools)
	}

	// Send request
	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(oaiReq).
		SetResult(&openAIChatResponse{}).
		Post("/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	oaiResp := resp.Result().(*openAIChatResponse)

	// Convert response
	return &ChatResponse{
		ID:    oaiResp.ID,
		Model: oaiResp.Model,
		Message: p.converter.FromOpenAIMessage(OpenAIMessage{
			Role:      oaiResp.Choices[0].Message.Role,
			Content:   oaiResp.Choices[0].Message.Content,
			ToolCalls: oaiResp.Choices[0].Message.ToolCalls,
		}),
		Usage: Usage{
			PromptTokens:     oaiResp.Usage.PromptTokens,
			CompletionTokens: oaiResp.Usage.CompletionTokens,
			TotalTokens:      oaiResp.Usage.TotalTokens,
		},
		Done: true,
	}, nil
}

// StreamChat sends a streaming chat request
func (p *OpenAIProvider) StreamChat(ctx context.Context, req *ChatRequest) (<-chan types.StreamChunk, error) {
	ch := make(chan types.StreamChunk, 100)

	// Build OpenAI request
	oaiReq := &openAIChatRequest{
		Model:       p.config.Model,
		Messages:    p.converter.ToOpenAIMessages(req.Messages),
		Temperature: p.config.Temperature,
		MaxTokens:   p.config.MaxTokens,
		Stream:      true,
	}

	if len(req.Tools) > 0 {
		oaiReq.Tools = ToOpenAITools(req.Tools)
	}

	// Create HTTP request
	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Send request
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Start streaming goroutine
	go func() {
		defer close(ch)
		defer httpResp.Body.Close()

		scanner := bufio.NewScanner(httpResp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Check for data prefix
			if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
				continue
			}

			// Extract data
			data := line[6:]

			// Check for end of stream
			if data == "[DONE]" {
				ch <- types.StreamChunk{Done: true}
				return
			}

			fmt.Println("response: ", string(data))
			// Parse SSE data
			var streamResp openAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				ch <- types.StreamChunk{Error: fmt.Errorf("failed to parse stream: %w", err)}
				return
			}

			// Send chunk
			if len(streamResp.Choices) > 0 {
				delta := streamResp.Choices[0].Delta
				chunk := types.StreamChunk{
					Content: delta.Content,
					Done:    streamResp.Choices[0].FinishReason == "stop",
				}

				if len(delta.ToolCalls) > 0 {
					tc := delta.ToolCalls[0]
					// Parse arguments JSON string to map
					var args map[string]interface{}
					if tc.Function.Arguments != "" {
						if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
							args = make(map[string]interface{})
						}
					} else {
						args = make(map[string]interface{})
					}
					chunk.ToolCall = &types.ToolCall{
						ID:        tc.ID,
						Name:      tc.Function.Name,
						Arguments: args,
					}
				}

				ch <- chunk
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- types.StreamChunk{Error: fmt.Errorf("stream error: %w", err)}
		}
	}()

	return ch, nil
}

// GetTools returns all registered tools
func (p *OpenAIProvider) GetTools() []types.ToolDefinition {
	return p.registry.GetAllTools()
}

// RegisterTool registers a tool
func (p *OpenAIProvider) RegisterTool(tool types.ToolDefinition, executor types.ToolExecutor) {
	p.registry.Register(tool, executor)
}

// Close closes the provider
func (p *OpenAIProvider) Close() error {
	return nil
}

// OpenAI request/response types

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Tools       []OpenAITool    `json:"tools,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      OpenAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string           `json:"role,omitempty"`
			Content   string           `json:"content,omitempty"`
			ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
