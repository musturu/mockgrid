// Package store defines interfaces and types for persistence.
package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Common errors for store implementations.
var (
	ErrNotFound = errors.New("record not found")
)

// MessageStatus represents the delivery status of a message.
// Based on SendGrid's delivery events.
type MessageStatus string

const (
	StatusProcessed MessageStatus = "processed" // Message accepted and ready for delivery
	StatusDelivered MessageStatus = "delivered" // Message delivered to receiving server
	StatusDeferred  MessageStatus = "deferred"  // Receiving server rejected temporarily
	StatusBounce    MessageStatus = "bounce"    // Permanent delivery failure (hard bounce)
	StatusBlocked   MessageStatus = "blocked"   // Temporary delivery failure (soft bounce)
	StatusDropped   MessageStatus = "dropped"   // Message dropped before sending
)

// Message represents a stored email message with its delivery status.
type Message struct {
	MsgID         string        `json:"msg_id"`
	FromEmail     string        `json:"from_email"`
	ToEmail       string        `json:"to_email"`
	Subject       string        `json:"subject"`
	HTMLBody      string        `json:"html_body,omitempty"`
	TextBody      string        `json:"text_body,omitempty"`
	Status        MessageStatus `json:"status"`
	SMTPResponse  string        `json:"smtp_response,omitempty"`
	Reason        string        `json:"reason,omitempty"`
	Timestamp     int64         `json:"timestamp"`
	LastEventTime int64         `json:"last_event_time,omitempty"`
	OpensCount    int           `json:"opens_count,omitempty"`
	ClicksCount   int           `json:"clicks_count,omitempty"`
}

// GetQuery defines query parameters for fetching messages.
type GetQuery struct {
	ID     string
	Status MessageStatus
	Limit  int
	Offset int
}

// MessageStore defines the interface for message persistence.
type MessageStore interface {
	// Save persists a message to the store.
	Save(msg *Message) error

	// Get retrieves messages based on query parameters.
	// If query.ID is set, returns a single message or ErrNotFound.
	Get(query GetQuery) ([]*Message, error)

	// Close releases any resources held by the store.
	Close() error
}

// GenerateMessageID creates a unique message ID using timestamp and random bytes.
func GenerateMessageID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return fmt.Sprintf("%d.%s", time.Now().UnixNano(), hex.EncodeToString(b)), nil
}
