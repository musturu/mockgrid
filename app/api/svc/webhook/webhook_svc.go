// Package webhook provides webhook management services.
package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mustur/mockgrid/app/api/store"
)

// ErrNotFound is returned when a webhook is not found
var ErrNotFound = errors.New("webhook not found")

// Service manages webhook configurations
type Service struct {
	store      store.WebhookStore
	dispatcher store.EventDispatcher
}

// NewService creates a new webhook service
func NewService(store store.WebhookStore, dispatcher store.EventDispatcher) *Service {
	return &Service{
		store:      store,
		dispatcher: dispatcher,
	}
}

// CreateWebhookRequest is the request body for creating a webhook (SendGrid format)
type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"` // e.g., ["processed", "delivered", "bounce", "deferred", "blocked", "dropped"]
	Secret string   `json:"secret,omitempty"`
}

// WebhookResponse is the response format for webhook endpoints (SendGrid format)
type WebhookResponse struct {
	ID       string   `json:"id"`
	URL      string   `json:"url"`
	Events   []string `json:"events"`
	Enabled  bool     `json:"enabled"`
	Secret   string   `json:"secret,omitempty"` // Only in responses when just created
	Created  int64    `json:"created,omitempty"`
	Modified int64    `json:"modified,omitempty"`
}

// ListResponse wraps the webhook list
type ListResponse struct {
	Result []*WebhookResponse `json:"result"`
}

// HandleListWebhooks handles GET /webhooks
func (s *Service) HandleListWebhooks(w http.ResponseWriter, r *http.Request) {
	hooks, err := s.store.List()
	if err != nil {
		slog.Error("failed to list webhooks", "err", err)
		http.Error(w, `{"error":"failed to list webhooks"}`, http.StatusInternalServerError)
		return
	}

	var resp ListResponse
	for _, hook := range hooks {
		resp.Result = append(resp.Result, webhookToResponse(hook))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleGetWebhooks handles GET /webhooks/{id}
func (s *Service) HandleGetWebhook(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path) // may be empty for list

	hook, err := s.store.Get(id)
	if err != nil {
		slog.Error("failed to get webhook(s)", "id", id, "err", err)
		http.Error(w, `{"error":"failed to get webhook(s)"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhookToResponse(hook))
}

// HandleCreateWebhook handles POST /webhooks
func (s *Service) HandleCreateWebhook(w http.ResponseWriter, r *http.Request) {

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Events) == 0 {
		http.Error(w, `{"error":"events array is required"}`, http.StatusBadRequest)
		return
	}

	config := &store.WebhookConfig{
		ID:      generateID(),
		URL:     req.URL,
		Enabled: true,
		Events:  req.Events,
		Secret:  req.Secret,
	}

	if err := s.store.Create(config); err != nil {
		slog.Error("failed to create webhook", "err", err)
		http.Error(w, `{"error":"failed to create webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := webhookToResponse(config)
	resp.Secret = req.Secret // Include secret in creation response
	json.NewEncoder(w).Encode(resp)
}

// HandleUpdateWebhook handles PUT /webhooks/{id}
func (s *Service) HandleUpdateWebhook(w http.ResponseWriter, r *http.Request) {

	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"webhook id is required"}`, http.StatusBadRequest)
		return
	}

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	hook, err := s.store.Get(id)
	if err != nil || hook == nil {
		http.Error(w, `{"error":"webhook not found"}`, http.StatusNotFound)
		return
	}
	//

	if req.URL != "" {
		hook.URL = req.URL
	}
	if len(req.Events) > 0 {
		hook.Events = req.Events
	}
	if req.Secret != "" {
		hook.Secret = req.Secret
	}

	if err := s.store.Update(hook); err != nil {
		slog.Error("failed to update webhook", "id", id, "err", err)
		http.Error(w, `{"error":"failed to update webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhookToResponse(hook))
}

// HandleDeleteWebhook handles DELETE /webhooks/{id}
func (s *Service) HandleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"webhook id is required"}`, http.StatusBadRequest)
		return
	}

	if err := s.store.Delete(id); err != nil {
		slog.Error("failed to delete webhook", "id", id, "err", err)
		if errors.Is(err, ErrNotFound) {
			http.Error(w, `{"error":"webhook not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"failed to delete webhook"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleWebhook handles POST /webhooks/{id}/toggle to enable/disable
func (s *Service) HandleToggleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"webhook id is required"}`, http.StatusBadRequest)
		return
	}

	hook, err := s.store.Get(id)
	if err != nil || hook == nil {
		http.Error(w, `{"error":"webhook not found"}`, http.StatusNotFound)
		return
	}

	hook.Enabled = !hook.Enabled
	if err := s.store.Update(hook); err != nil {
		slog.Error("failed to toggle webhook", "id", id, "err", err)
		http.Error(w, `{"error":"failed to toggle webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhookToResponse(hook))
}

// Helper functions

func webhookToResponse(hook *store.WebhookConfig) *WebhookResponse {
	return &WebhookResponse{
		ID:      hook.ID,
		URL:     hook.URL,
		Events:  hook.Events,
		Enabled: hook.Enabled,
	}
}

func extractID(path string) string {
	// Parse /webhooks/{id} or /webhooks/{id}/toggle
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

func generateID() string {
	// Simple ID generation (in production, use uuid)
	return fmt.Sprintf("wh_%d", int64(len([]byte{})))
}
