// Package store defines interfaces and types for persistence.
package store

// NoOpDispatcher is a no-op implementation of EventDispatcher that discards all events.
type NoOpDispatcher struct{}

// DispatchMessageEvent discards the event and returns immediately.
func (n *NoOpDispatcher) DispatchMessageEvent(msgID, email, from, subject string, status string, reason string) {
	// no-op
}
