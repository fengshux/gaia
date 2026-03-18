// Package storage provides data persistence
package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	"gaia/pkg/types"
	"gaia/pkg/utils"
)

// SessionRepository handles session persistence
type SessionRepository struct {
	db *SQLiteStorage
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *SQLiteStorage) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(session *types.Session) error {
	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO sessions (id, name, created_at, updated_at, metadata)
		 VALUES (?, ?, ?, ?, ?)`,
		session.ID,
		session.Name,
		session.CreatedAt,
		session.UpdatedAt,
		string(metadata),
	)

	return err
}

// Get retrieves a session by ID
func (r *SessionRepository) Get(id string) (*types.Session, error) {
	var session types.Session
	var metadata string

	err := r.db.QueryRow(
		`SELECT id, name, created_at, updated_at, metadata
		 FROM sessions WHERE id = ?`,
		id,
	).Scan(
		&session.ID,
		&session.Name,
		&session.CreatedAt,
		&session.UpdatedAt,
		&metadata,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if metadata != "" {
		json.Unmarshal([]byte(metadata), &session.Metadata)
	}

	return &session, nil
}

// Update updates a session
func (r *SessionRepository) Update(session *types.Session) error {
	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return err
	}

	session.UpdatedAt = time.Now()

	_, err = r.db.Exec(
		`UPDATE sessions SET name = ?, updated_at = ?, metadata = ? WHERE id = ?`,
		session.Name,
		session.UpdatedAt,
		string(metadata),
		session.ID,
	)

	return err
}

// Delete removes a session
func (r *SessionRepository) Delete(id string) error {
	// Delete messages first
	_, err := r.db.Exec("DELETE FROM messages WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	// Delete session
	_, err = r.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

// List returns all sessions
func (r *SessionRepository) List() ([]*types.Session, error) {
	rows, err := r.db.Query(
		`SELECT id, name, created_at, updated_at, metadata
		 FROM sessions ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*types.Session
	for rows.Next() {
		var session types.Session
		var metadata string

		err := rows.Scan(
			&session.ID,
			&session.Name,
			&session.CreatedAt,
			&session.UpdatedAt,
			&metadata,
		)
		if err != nil {
			return nil, err
		}

		if metadata != "" {
			json.Unmarshal([]byte(metadata), &session.Metadata)
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// GetOrCreate gets a session or creates a new one
func (r *SessionRepository) GetOrCreate(id string) (*types.Session, error) {
	if id != "" {
		session, err := r.Get(id)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	// Create new session
	session := &types.Session{
		ID:        utils.GenerateID(),
		Name:      "New Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []types.Message{},
		Metadata:  make(map[string]interface{}),
	}

	if err := r.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}
