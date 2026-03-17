// Package llm provides LLM provider implementations
package llm

import (
	"assistant/pkg/types"
)

// MessageConverter converts between internal and OpenAI message formats
type MessageConverter struct{}

// ToOpenAIMessage converts internal message to OpenAI format
func (mc *MessageConverter) ToOpenAIMessage(msg types.Message) OpenAIMessage {
	oaiMsg := OpenAIMessage{
		Role:    string(msg.Role),
		Content: msg.Content,
	}

	if len(msg.ToolCalls) > 0 {
		oaiMsg.ToolCalls = make([]OpenAIToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			oaiMsg.ToolCalls[i] = OpenAIToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: OpenAIFunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			}
		}
	}

	if msg.ToolID != "" {
		oaiMsg.ToolCallID = msg.ToolID
	}

	return oaiMsg
}

// ToOpenAIMessages converts internal messages to OpenAI format
func (mc *MessageConverter) ToOpenAIMessages(msgs []types.Message) []OpenAIMessage {
	result := make([]OpenAIMessage, len(msgs))
	for i, msg := range msgs {
		result[i] = mc.ToOpenAIMessage(msg)
	}
	return result
}

// FromOpenAIMessage converts OpenAI message to internal format
func (mc *MessageConverter) FromOpenAIMessage(msg OpenAIMessage) types.Message {
	result := types.Message{
		Role:    types.Role(msg.Role),
		Content: msg.Content,
	}

	if len(msg.ToolCalls) > 0 {
		result.ToolCalls = make([]types.ToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			result.ToolCalls[i] = types.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}

	if msg.ToolCallID != "" {
		result.ToolID = msg.ToolCallID
	}

	return result
}

// OpenAIMessage represents an OpenAI message
type OpenAIMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content"`
	ToolCalls  []OpenAIToolCall       `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Name       string                 `json:"name,omitempty"`
}

// OpenAIToolCall represents an OpenAI tool call
type OpenAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function OpenAIFunctionCall `json:"function"`
}

// OpenAIFunctionCall represents an OpenAI function call
type OpenAIFunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// OpenAITool represents an OpenAI tool definition
type OpenAITool struct {
	Type     string              `json:"type"`
	Function OpenAIToolFunction  `json:"function"`
}

// OpenAIToolFunction represents an OpenAI tool function
type OpenAIToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ToOpenAITools converts tool definitions to OpenAI format
func ToOpenAITools(tools []types.ToolDefinition) []OpenAITool {
	result := make([]OpenAITool, len(tools))
	for i, t := range tools {
		result[i] = OpenAITool{
			Type: "function",
			Function: OpenAIToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return result
}
