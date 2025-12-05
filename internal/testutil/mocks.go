// Package testutil provides test helpers and mock implementations.
package testutil

import (
	"net/http"
	"sync"

	"github.com/mustur/mockgrid/app/api/middleware"
	"github.com/mustur/mockgrid/app/api/store"
	"github.com/mustur/mockgrid/app/template"
)

// --- MockService implements api.Service for testing ---

// MockService is a configurable mock implementation of api.Service.
type MockService struct {
	mux   *http.ServeMux
	root  string
	chain middleware.Middleware
}

// NewMockService creates a MockService with the given root path.
func NewMockService(root string) *MockService {
	return &MockService{
		mux:   http.NewServeMux(),
		root:  root,
		chain: middleware.Chain(), // identity chain by default
	}
}

// GetMux returns the service's ServeMux.
func (m *MockService) GetMux() *http.ServeMux {
	return m.mux
}

// GetRoot returns the service's root path.
func (m *MockService) GetRoot() string {
	return m.root
}

// Chain returns the service's middleware chain.
func (m *MockService) Chain() middleware.Middleware {
	return m.chain
}

// WithChain sets the middleware chain and returns the service for chaining.
func (m *MockService) WithChain(mw middleware.Middleware) *MockService {
	m.chain = mw
	return m
}

// Handle registers a handler for the given pattern.
func (m *MockService) Handle(pattern string, handler http.Handler) *MockService {
	m.mux.Handle(pattern, handler)
	return m
}

// HandleFunc registers a handler function for the given pattern.
func (m *MockService) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *MockService {
	m.mux.HandleFunc(pattern, handler)
	return m
}

// --- RecordingMiddleware captures middleware execution ---

// MiddlewareRecord captures information about a middleware invocation.
type MiddlewareRecord struct {
	Name   string
	Method string
	Path   string
	Phase  string // "before" or "after"
}

// RecordingMiddleware creates a middleware that records its execution.
type RecordingMiddleware struct {
	mu      sync.Mutex
	records []MiddlewareRecord
}

// NewRecordingMiddleware creates a new RecordingMiddleware.
func NewRecordingMiddleware() *RecordingMiddleware {
	return &RecordingMiddleware{}
}

// Middleware returns a named middleware function that records execution.
func (r *RecordingMiddleware) Middleware(name string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r.record(name, req.Method, req.URL.Path, "before")
			next.ServeHTTP(w, req)
			r.record(name, req.Method, req.URL.Path, "after")
		})
	}
}

func (r *RecordingMiddleware) record(name, method, path, phase string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, MiddlewareRecord{
		Name:   name,
		Method: method,
		Path:   path,
		Phase:  phase,
	})
}

// Records returns a copy of all recorded invocations.
func (r *RecordingMiddleware) Records() []MiddlewareRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]MiddlewareRecord, len(r.records))
	copy(cp, r.records)
	return cp
}

// Reset clears all recorded invocations.
func (r *RecordingMiddleware) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = nil
}

// --- MockTemplater implements template.Templater ---

// MockTemplater is a configurable mock for template.Templater.
type MockTemplater struct {
	Templates map[string]*template.TemplateVersion
	Error     error
}

// NewMockTemplater creates an empty MockTemplater.
func NewMockTemplater() *MockTemplater {
	return &MockTemplater{
		Templates: make(map[string]*template.TemplateVersion),
	}
}

// GetTemplate returns a template by ID or the configured error.
func (m *MockTemplater) GetTemplate(templateID string) (*template.TemplateVersion, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if tmpl, ok := m.Templates[templateID]; ok {
		return tmpl, nil
	}
	return nil, nil
}

// WithTemplate adds a template and returns the mock for chaining.
func (m *MockTemplater) WithTemplate(id string, v *template.TemplateVersion) *MockTemplater {
	m.Templates[id] = v
	return m
}

// WithError sets the error to return and returns the mock for chaining.
func (m *MockTemplater) WithError(err error) *MockTemplater {
	m.Error = err
	return m
}

// --- MockMessageStore implements store.MessageStore ---

// MockMessageStore is an in-memory mock for store.MessageStore.
type MockMessageStore struct {
	mu       sync.Mutex
	messages map[string]*store.Message
	SaveErr  error
	GetErr   error
}

// NewMockMessageStore creates an empty MockMessageStore.
func NewMockMessageStore() *MockMessageStore {
	return &MockMessageStore{
		messages: make(map[string]*store.Message),
	}
}

// Save stores a message in memory.
func (m *MockMessageStore) Save(msg *store.Message) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *msg
	m.messages[msg.MsgID] = &cp
	return nil
}

// Get retrieves messages matching the query.
func (m *MockMessageStore) Get(q store.GetQuery) ([]*store.Message, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if q.ID != "" {
		if msg, ok := m.messages[q.ID]; ok {
			cp := *msg
			return []*store.Message{&cp}, nil
		}
		return []*store.Message{}, nil
	}

	var result []*store.Message
	for _, msg := range m.messages {
		if q.Status != "" && msg.Status != q.Status {
			continue
		}
		cp := *msg
		result = append(result, &cp)
	}

	if q.Limit > 0 && len(result) > q.Limit {
		result = result[:q.Limit]
	}
	return result, nil
}

// Close is a no-op for the mock.
func (m *MockMessageStore) Close() error {
	return nil
}

// Messages returns all stored messages (for test assertions).
func (m *MockMessageStore) Messages() []*store.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*store.Message, 0, len(m.messages))
	for _, msg := range m.messages {
		cp := *msg
		result = append(result, &cp)
	}
	return result
}

// Reset clears all stored messages.
func (m *MockMessageStore) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make(map[string]*store.Message)
}
