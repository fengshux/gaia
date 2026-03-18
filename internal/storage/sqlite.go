// Package storage provides data persistence
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gaia/pkg/utils"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage provides SQLite-based storage
type SQLiteStorage struct {
	mu   sync.RWMutex
	db   *sql.DB
	path string
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := utils.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	storage := &SQLiteStorage{
		db:   db,
		path: path,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema initializes the database schema
func (s *SQLiteStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		metadata TEXT
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		tool_calls TEXT,
		tool_id TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying database connection
func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}

// Exec executes a query
func (s *SQLiteStorage) Exec(query string, args ...interface{}) (sql.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Exec(query, args...)
}

// Query executes a query that returns rows
func (s *SQLiteStorage) Query(query string, args ...interface{}) (*sql.Rows, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Query(query, args...)
}

// QueryRow executes a query that returns at most one row
func (s *SQLiteStorage) QueryRow(query string, args ...interface{}) *sql.Row {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.QueryRow(query, args...)
}

// GetSetting retrieves a setting by key
func (s *SQLiteStorage) GetSetting(key string) (string, error) {
	var value string
	err := s.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting stores a setting
func (s *SQLiteStorage) SetSetting(key, value string) error {
	_, err := s.Exec(
		"INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

// DeleteSetting removes a setting
func (s *SQLiteStorage) DeleteSetting(key string) error {
	_, err := s.Exec("DELETE FROM settings WHERE key = ?", key)
	return err
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
