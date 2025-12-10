// Package store provides message storage interfaces.
package store

import (
	"log/slog"
)

// StoreWrapper wraps a MessageStore and dispatches webhooks on status changes
type StoreWrapper struct {
	wrapped    MessageStore
	dispatcher EventDispatcher
}

// NewStoreWrapper creates a new wrapper.
// dispatcher must not be nil; use NoOpDispatcher if no event dispatch is desired.
func NewStoreWrapper(baseStore MessageStore, dispatcher EventDispatcher) *StoreWrapper {
	return &StoreWrapper{
		wrapped:    baseStore,
		dispatcher: dispatcher,
	}
}

// Save persists a message and dispatches webhook if status changed
func (w *StoreWrapper) Save(msg *Message) error {
	// Check if this is an update (message already exists)
	oldMsgs, _ := w.wrapped.Get(GetQuery{ID: msg.MsgID})
	wasNew := len(oldMsgs) == 0

	// Save to underlying store
	if err := w.wrapped.Save(msg); err != nil {
		return err
	}

	// Dispatch webhook for new message or status change
	if wasNew || (len(oldMsgs) > 0 && oldMsgs[0].Status != msg.Status) {
		slog.Debug("dispatching webhook event", "msg_id", msg.MsgID, "status", msg.Status)
		w.dispatcher.DispatchMessageEvent(
			msg.MsgID,
			msg.ToEmail,
			msg.FromEmail,
			msg.Subject,
			string(msg.Status),
			msg.Reason,
		)
	}

	return nil
}

// Get delegates to wrapped store
func (w *StoreWrapper) Get(query GetQuery) ([]*Message, error) {
	return w.wrapped.Get(query)
}

// Close delegates to wrapped store
func (w *StoreWrapper) Close() error {
	return w.wrapped.Close()
}
