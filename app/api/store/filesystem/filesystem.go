// Package filesystem provides a file-based message store implementation.
package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mustur/mockgrid/app/api/store"
)

// Store persists messages as individual JSON files.
type Store struct {
	dir string
}

// New creates a new filesystem store at the given directory.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a message to a JSON file named by its ID.
func (s *Store) Save(msg *store.Message) error {
	if msg.MsgID == "" {
		return fmt.Errorf("message ID is required")
	}

	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	filename := s.filename(msg.MsgID)
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("write message file: %w", err)
	}

	return nil
}

// Get retrieves messages based on query parameters.
func (s *Store) Get(query store.GetQuery) ([]*store.Message, error) {
	if query.ID != "" {
		return s.getByID(query.ID)
	}
	return s.list(query)
}

// Close is a no-op for filesystem store.
func (s *Store) Close() error {
	return nil
}

func (s *Store) getByID(id string) ([]*store.Message, error) {
	filename := s.filename(id)

	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read message file: %w", err)
	}

	var msg store.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}

	return []*store.Message{&msg}, nil
}

func (s *Store) list(query store.GetQuery) ([]*store.Message, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read store directory: %w", err)
	}

	limit := query.Limit
	if limit == 0 {
		limit = 100
	}

	var messages []*store.Message
	skipped := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		msg, err := s.readMessageFile(entry.Name())
		if err != nil {
			continue
		}

		if query.Status != "" && msg.Status != query.Status {
			continue
		}

		if skipped < query.Offset {
			skipped++
			continue
		}

		messages = append(messages, msg)
		if len(messages) >= limit {
			break
		}
	}

	return messages, nil
}

func (s *Store) readMessageFile(name string) (*store.Message, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name))
	if err != nil {
		return nil, err
	}

	var msg store.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (s *Store) filename(id string) string {
	safeID := filepath.Base(id)
	return filepath.Join(s.dir, safeID+".json")
}
