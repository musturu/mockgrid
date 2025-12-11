package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// Store persists messages as individual JSON files.
type Store struct {
	dir string
}

func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Close is a no-op for filesystem store.
func (s *Store) Close() error {
	return nil
}

// Connect is a no-op for filesystem store.
func (s *Store) Connect() error {
	return nil
}

func (s *Store) filename(id string) string {
	safeID := filepath.Base(id)
	return filepath.Join(s.dir, safeID+".json")
}
