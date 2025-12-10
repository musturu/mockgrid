package store

// WebhookConfig holds webhook registration data
type WebhookConfig struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Enabled   bool     `json:"enabled"`
	Events    []string `json:"events"` // event types to send
	Secret    string   `json:"secret,omitempty"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

// WebhookStore defines persistence for webhook configurations
type WebhookStore interface {
	// Create stores a new webhook
	Create(hook *WebhookConfig) error

	// Get retrieves webhooks based on query parameters.
	Get(id string) (*WebhookConfig, error)

	List() ([]*WebhookConfig, error)

	// ListEnabled lists all enabled webhooks
	ListEnabled() ([]*WebhookConfig, error)

	// Update modifies a webhook
	Update(hook *WebhookConfig) error

	// Delete removes a webhook
	Delete(id string) error

	// Close releases resources
	Close() error
}
