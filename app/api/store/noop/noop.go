// Package noop provides a no-op message store implementation.
package noop

import (
	"github.com/mustur/mockgrid/app/api/store"
)

// Store is a no-op implementation of MessageStore.
// It discards all messages and returns empty results.
type Store struct{}

// New creates a new no-op store.
func New() *Store {
	return &Store{}
}

// Save discards the message and returns nil.
func (s *Store) Save(_ *store.Message) error {
	return nil
}

// Get always returns an empty slice.
func (s *Store) Get(_ store.GetQuery) ([]*store.Message, error) {
	return []*store.Message{}, nil
}

// Close is a no-op.
func (s *Store) Close() error {
	return nil
}
