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

	// GetWebhook retrieves a webhook by ID
	GetWebhook(id string) (*WebhookConfig, error)

	ListWebhooks() ([]*WebhookConfig, error)

	// ListEnabledWebhooks lists all enabled webhooks
	ListEnabledWebhooks() ([]*WebhookConfig, error)

	// UpdateWebhook modifies a webhook
	UpdateWebhook(hook *WebhookConfig) error

	// DeleteWebhook removes a webhook by ID
	DeleteWebhook(id string) error

	// Close releases resources
	Close() error
}
