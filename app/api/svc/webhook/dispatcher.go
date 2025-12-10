// Package webhook provides webhook management and event dispatching.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mustur/mockgrid/app/api/store"
)

// Dispatcher sends webhook events to registered endpoints.
// It implements store.EventDispatcher.
type Dispatcher struct {
	webhookStore store.WebhookStore
	httpClient   *http.Client
}

// Event represents a webhook event payload (SendGrid format)
type Event struct {
	EventID   string `json:"event_id"`
	Type      string `json:"event"` // "processed", "delivered", "bounce", etc.
	Timestamp int64  `json:"timestamp"`
	MessageID string `json:"sg_message_id"`
	Email     string `json:"email"`
	From      string `json:"from,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Status    string `json:"status,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher(store store.WebhookStore) *Dispatcher {
	return &Dispatcher{
		webhookStore: store,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DispatchMessageEvent sends an event to all registered webhooks that match the event type
// This runs in a goroutine to avoid blocking the caller
func (d *Dispatcher) DispatchMessageEvent(msgID, email, from, subject string, status string, reason string) {
	go d.dispatchAsync(msgID, email, from, subject, status, reason)
}

func (d *Dispatcher) dispatchAsync(msgID, email, from, subject string, status string, reason string) {
	// Get all enabled webhooks
	webhooks, err := d.webhookStore.ListEnabled()
	if err != nil {
		slog.Error("failed to list webhooks", "err", err)
		return
	}

	if len(webhooks) == 0 {
		slog.Debug("no webhooks registered, skipping dispatch")
		return
	}

	for _, hook := range webhooks {
		// Check if this webhook is subscribed to this event type
		if !isSubscribed(hook, status) {
			slog.Debug("webhook not subscribed to event",
				"webhook_id", hook.ID, "event_type", status)
			continue
		}

		// Send to this webhook with retries
		d.sendWithRetry(hook, msgID, email, from, subject, status, reason)
	}
}

// sendWithRetry sends an event with exponential backoff retries
func (d *Dispatcher) sendWithRetry(hook *store.WebhookConfig, msgID, email, from, subject, status, reason string) {
	maxRetries := 3
	backoff := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := d.send(hook, msgID, email, from, subject, status, reason); err == nil {
			slog.Info("webhook delivered", "webhook_id", hook.ID, "event_type", status)
			return
		} else {
			slog.Warn("webhook delivery failed",
				"webhook_id", hook.ID,
				"attempt", attempt+1,
				"err", err)
		}

		if attempt < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2 // exponential backoff
		}
	}

	slog.Error("webhook delivery failed after retries",
		"webhook_id", hook.ID,
		"event_type", status)
}

// send delivers the event to a single webhook endpoint
func (d *Dispatcher) send(hook *store.WebhookConfig, msgID, email, from, subject, status, reason string) error {
	event := &Event{
		EventID:   fmt.Sprintf("%d-%s", time.Now().UnixNano(), msgID),
		Type:      status,
		Timestamp: time.Now().Unix(),
		MessageID: msgID,
		Email:     email,
		From:      from,
		Subject:   subject,
		Status:    status,
		Reason:    reason,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mockgrid/1.0")

	// Add HMAC signature if secret is configured
	if hook.Secret != "" {
		signature := d.generateSignature(payload, hook.Secret)
		req.Header.Set("X-Twilio-Signature", signature)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Consume response body to allow connection reuse
	_, _ = io.ReadAll(resp.Body)

	// Only accept 2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// generateSignature creates an HMAC-SHA256 signature for the payload (SendGrid style)
func (d *Dispatcher) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// Helper to check if webhook is subscribed to event type
func isSubscribed(hook *store.WebhookConfig, eventType string) bool {
	for _, e := range hook.Events {
		if e == eventType {
			return true
		}
	}
	return false
}
