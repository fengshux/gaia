// Package core provides the core engine
package core

import (
	"sync"
	"time"

	"gaia/pkg/types"
	"gaia/pkg/utils"
)

// SessionManager manages conversation sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*types.Session
	current  string
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*types.Session),
	}
}

// Create creates a new session
func (sm *SessionManager) Create(name string) *types.Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &types.Session{
		ID:        utils.GenerateID(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []types.Message{},
		Metadata:  make(map[string]interface{}),
	}

	sm.sessions[session.ID] = session

	if sm.current == "" {
		sm.current = session.ID
	}

	return session
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(id string) (*types.Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[id]
	return session, ok
}

// GetOrCreate gets a session or creates a new one if ID is empty
func (sm *SessionManager) GetOrCreate(id string) *types.Session {
	if id == "" {
		id = sm.current
	}

	if id == "" {
		return sm.Create("New Session")
	}

	session, ok := sm.Get(id)
	if !ok {
		return sm.Create("New Session")
	}

	return session
}

// GetCurrent returns the current session
func (sm *SessionManager) GetCurrent() *types.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.current == "" {
		return nil
	}

	return sm.sessions[sm.current]
}

// SetCurrent sets the current session
func (sm *SessionManager) SetCurrent(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.sessions[id]; ok {
		sm.current = id
		return true
	}
	return false
}

// Delete removes a session
func (sm *SessionManager) Delete(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.sessions[id]; ok {
		delete(sm.sessions, id)
		if sm.current == id {
			sm.current = ""
		}
		return true
	}
	return false
}

// List returns all sessions
func (sm *SessionManager) List() []*types.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*types.Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// AddMessage adds a message to a session
func (sm *SessionManager) AddMessage(sessionID string, msg types.Message) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return false
	}

	msg.ID = utils.GenerateID()
	msg.CreatedAt = time.Now()
	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = time.Now()

	return true
}
