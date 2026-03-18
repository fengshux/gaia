// Package storage provides data persistence
package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	"gaia/pkg/types"
	"gaia/pkg/utils"
)

// MessageRepository handles message persistence
type MessageRepository struct {
	db *SQLiteStorage
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *SQLiteStorage) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(sessionID string, message *types.Message) error {
	toolCalls, err := json.Marshal(message.ToolCalls)
	if err != nil {
		return err
	}

	if message.ID == "" {
		message.ID = utils.GenerateID()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	_, err = r.db.Exec(
		`INSERT INTO messages (id, session_id, role, content, tool_calls, tool_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		message.ID,
		sessionID,
		message.Role,
		message.Content,
		string(toolCalls),
		message.ToolID,
		message.CreatedAt,
	)

	return err
}

// Get retrieves a message by ID
func (r *MessageRepository) Get(id string) (*types.Message, error) {
	var message types.Message
	var toolCalls string

	err := r.db.QueryRow(
		`SELECT id, role, content, tool_calls, tool_id, created_at
		 FROM messages WHERE id = ?`,
		id,
	).Scan(
		&message.ID,
		&message.Role,
		&message.Content,
		&toolCalls,
		&message.ToolID,
		&message.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if toolCalls != "" {
		json.Unmarshal([]byte(toolCalls), &message.ToolCalls)
	}

	return &message, nil
}

// GetBySession retrieves all messages for a session
func (r *MessageRepository) GetBySession(sessionID string) ([]types.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, role, content, tool_calls, tool_id, created_at
		 FROM messages WHERE session_id = ? ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var message types.Message
		var toolCalls string

		err := rows.Scan(
			&message.ID,
			&message.Role,
			&message.Content,
			&toolCalls,
			&message.ToolID,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if toolCalls != "" {
			json.Unmarshal([]byte(toolCalls), &message.ToolCalls)
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// Delete removes a message
func (r *MessageRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM messages WHERE id = ?", id)
	return err
}

// DeleteBySession removes all messages for a session
func (r *MessageRepository) DeleteBySession(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	return err
}

// GetRecent retrieves recent messages for a session
func (r *MessageRepository) GetRecent(sessionID string, limit int) ([]types.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, role, content, tool_calls, tool_id, created_at
		 FROM messages WHERE session_id = ?
		 ORDER BY created_at DESC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var message types.Message
		var toolCalls string

		err := rows.Scan(
			&message.ID,
			&message.Role,
			&message.Content,
			&toolCalls,
			&message.ToolID,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if toolCalls != "" {
			json.Unmarshal([]byte(toolCalls), &message.ToolCalls)
		}

		messages = append(messages, message)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Count returns the number of messages for a session
func (r *MessageRepository) Count(sessionID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE session_id = ?",
		sessionID,
	).Scan(&count)
	return count, err
}
