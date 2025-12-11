package store

type Storer interface {
	Close() error
	Connect() error
}

// BackendStore combines message and webhook stores and provides connection lifecycle
type BackendStore interface {
	MessageStore
	WebhookStore
	Storer
}
