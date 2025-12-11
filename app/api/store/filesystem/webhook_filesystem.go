package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mustur/mockgrid/app/api/store"
)

// webhookFile returns a safe path for storing webhook with the given id.
func (s *Store) webhookFile(id string) string {
	safe := filepath.Base(id)
	return filepath.Join(s.dir, "webhooks", safe+".json")
}

func (s *Store) ensureWebhookDir() error {
	return os.MkdirAll(filepath.Join(s.dir, "webhooks"), 0o750)
}

// Create saves a webhook definition to disk.
func (s *Store) Create(hook *store.WebhookConfig) error {
	if hook.ID == "" {
		return errors.New("webhook id is required")
	}
	if err := s.ensureWebhookDir(); err != nil {
		return err
	}
	file := s.webhookFile(hook.ID)
	if _, err := os.Stat(file); err == nil {
		return errors.New("webhook already exists")
	}
	if hook.CreatedAt == 0 {
		hook.CreatedAt = time.Now().Unix()
	}
	if hook.UpdatedAt == 0 {
		hook.UpdatedAt = hook.CreatedAt
	}
	data, err := json.MarshalIndent(hook, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, data, 0o600); err != nil {
		return fmt.Errorf("write webhook file: %w", err)
	}
	return nil
}

// GetWebhook reads a webhook by id.
func (s *Store) GetWebhook(id string) (*store.WebhookConfig, error) {
	file := filepath.Base(id)
	fsys := os.DirFS(filepath.Join(s.dir, "webhooks"))
	data, err := fs.ReadFile(fsys, file+".json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	var cfg store.WebhookConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ListWebhooks returns all registered webhooks.
func (s *Store) ListWebhooks() ([]*store.WebhookConfig, error) {
	dir := filepath.Join(s.dir, "webhooks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	fsys := os.DirFS(dir)
	var res []*store.WebhookConfig
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			continue
		}
		var cfg store.WebhookConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		res = append(res, &cfg)
	}
	return res, nil
}

// ListEnabledWebhooks returns all enabled webhooks.
func (s *Store) ListEnabledWebhooks() ([]*store.WebhookConfig, error) {
	all, err := s.ListWebhooks()
	if err != nil {
		return nil, err
	}
	var res []*store.WebhookConfig
	for _, w := range all {
		if w.Enabled {
			res = append(res, w)
		}
	}
	return res, nil
}

// UpdateWebhook updates an existing webhook file.
func (s *Store) UpdateWebhook(hook *store.WebhookConfig) error {
	file := s.webhookFile(hook.ID)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return store.ErrNotFound
	}
	hook.UpdatedAt = time.Now().Unix()
	data, err := json.MarshalIndent(hook, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o600)
}

// DeleteWebhook removes a webhook file.
func (s *Store) DeleteWebhook(id string) error {
	file := s.webhookFile(id)
	if err := os.Remove(file); err != nil {
		if os.IsNotExist(err) {
			return store.ErrNotFound
		}
		return err
	}
	return nil
}
