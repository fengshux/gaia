// Package core provides the core engine
package core

import (
	"context"
	"strings"
	"time"

	"gaia/pkg/types"
)

// ContextManager manages conversation context
type ContextManager struct {
	maxMessages int
	maxTokens   int
}

// NewContextManager creates a new context manager
func NewContextManager(maxMessages, maxTokens int) *ContextManager {
	return &ContextManager{
		maxMessages: maxMessages,
		maxTokens:   maxTokens,
	}
}

// BuildContext builds the message context for a session
func (cm *ContextManager) BuildContext(session *types.Session, systemPrompt string) []types.Message {
	messages := make([]types.Message, 0)

	// Add system prompt if provided
	if systemPrompt != "" {
		messages = append(messages, types.Message{
			Role:    types.RoleSystem,
			Content: systemPrompt,
		})
	}

	// Add conversation messages
	// Limit to max messages
	start := 0
	if len(session.Messages) > cm.maxMessages {
		start = len(session.Messages) - cm.maxMessages
	}

	for i := start; i < len(session.Messages); i++ {
		messages = append(messages, session.Messages[i])
	}

	return messages
}

// EstimateTokens estimates the number of tokens in a message
func (cm *ContextManager) EstimateTokens(msg types.Message) int {
	// Simple estimation: ~4 characters per token
	content := msg.Content
	for _, tc := range msg.ToolCalls {
		content += tc.Name
		for k, v := range tc.Arguments {
			content += k + strings.Repeat("x", 10) // Simplified
			_ = v
		}
	}
	return len(content) / 4
}

// TrimMessages trims messages to fit within token limit
func (cm *ContextManager) TrimMessages(messages []types.Message, maxTokens int) []types.Message {
	if maxTokens <= 0 {
		maxTokens = cm.maxTokens
	}

	// Always keep system message if present
	result := make([]types.Message, 0)
	totalTokens := 0

	// Check if first message is system
	startIdx := 0
	if len(messages) > 0 && messages[0].Role == types.RoleSystem {
		tokens := cm.EstimateTokens(messages[0])
		result = append(result, messages[0])
		totalTokens += tokens
		startIdx = 1
	}

	// Add messages from newest to oldest (but keep order)
	for i := startIdx; i < len(messages); i++ {
		tokens := cm.EstimateTokens(messages[i])
		if totalTokens+tokens > maxTokens {
			break
		}
		result = append(result, messages[i])
		totalTokens += tokens
	}

	return result
}

// SummarizeContext creates a summary of old messages
func (cm *ContextManager) SummarizeContext(ctx context.Context, messages []types.Message) string {
	// This would typically call an LLM to summarize
	// For now, return a simple summary
	var summary strings.Builder
	summary.WriteString("Previous conversation summary:\n")

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleUser:
			summary.WriteString("User asked: ")
		case types.RoleAssistant:
			summary.WriteString("Assistant responded: ")
		}
		// Truncate content
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		summary.WriteString(content + "\n")
	}

	return summary.String()
}

// Conversation represents a conversation with metadata
type Conversation struct {
	ID        string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []types.Message
	Summary   string
	Tags      []string
}

// ConversationManager manages conversations
type ConversationManager struct {
	conversations map[string]*Conversation
}

// NewConversationManager creates a new conversation manager
func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		conversations: make(map[string]*Conversation),
	}
}

// Create creates a new conversation
func (cm *ConversationManager) Create(title string) *Conversation {
	conv := &Conversation{
		ID:        generateID(),
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []types.Message{},
		Tags:      []string{},
	}
	cm.conversations[conv.ID] = conv
	return conv
}

// Get retrieves a conversation by ID
func (cm *ConversationManager) Get(id string) (*Conversation, bool) {
	conv, ok := cm.conversations[id]
	return conv, ok
}

// Delete removes a conversation
func (cm *ConversationManager) Delete(id string) {
	delete(cm.conversations, id)
}

// List returns all conversations
func (cm *ConversationManager) List() []*Conversation {
	convs := make([]*Conversation, 0, len(cm.conversations))
	for _, c := range cm.conversations {
		convs = append(convs, c)
	}
	return convs
}

// Helper function
func generateID() string {
	return time.Now().Format("20060102150405")
}
