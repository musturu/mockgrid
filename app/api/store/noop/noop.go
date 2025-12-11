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

// SaveMSG discards the message and returns nil.
func (s *Store) SaveMSG(_ *store.Message) error {
	return nil
}

// GetMSG always returns an empty slice.
func (s *Store) GetMSG(_ store.GetQuery) ([]*store.Message, error) {
	return []*store.Message{}, nil
}

// Close is a no-op.
func (s *Store) Close() error {
	return nil
}

// Connect is a no-op.
func (s *Store) Connect() error {
	return nil
}

// Webhook store no-op implementations
func (s *Store) Create(_ *store.WebhookConfig) error               { return nil }
func (s *Store) GetWebhook(_ string) (*store.WebhookConfig, error) { return nil, store.ErrNotFound }
func (s *Store) ListWebhooks() ([]*store.WebhookConfig, error)     { return []*store.WebhookConfig{}, nil }
func (s *Store) ListEnabledWebhooks() ([]*store.WebhookConfig, error) {
	return []*store.WebhookConfig{}, nil
}
func (s *Store) UpdateWebhook(_ *store.WebhookConfig) error { return store.ErrNotFound }
func (s *Store) DeleteWebhook(_ string) error               { return store.ErrNotFound }
