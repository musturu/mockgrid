// Package store defines interfaces and types for persistence.
package store

// EventDispatcher dispatches events (e.g., webhooks) when message status changes.
// Implementations should handle event delivery asynchronously to avoid blocking message operations.
type EventDispatcher interface {
	// DispatchMessageEvent is called when a message status changes.
	// Implementations should not block the caller.
	DispatchMessageEvent(msgID, email, from, subject string, status string, reason string)
}
