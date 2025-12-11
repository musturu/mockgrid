package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mustur/mockgrid/app/api/middleware"
)

// Service defines the interface that all services must implement.
type Service interface {
	GetMux() *http.ServeMux
	GetRoot() string
	Chain() middleware.Middleware
}

// MockGrid is the main application server.
type MockGrid struct {
	services   []Service
	listenAddr string
}

// New creates a new MockGrid instance with the given services.
func New(listenAddr string, services ...Service) *MockGrid {
	return &MockGrid{
		listenAddr: listenAddr,
		services:   services,
	}
}

// Start initializes and starts the HTTP server.
func (m *MockGrid) Start() error {
	if len(m.services) == 0 {
		return errors.New("no services registered")
	}

	mux := http.NewServeMux()

	for _, svc := range m.services {
		root := svc.GetRoot()
		handler := svc.Chain()(svc.GetMux())
		// StripPrefix needs the path without trailing slash to avoid redirect issues
		// e.g., /api/ -> strip /api so /api/test becomes /test (not redirect to /test)
		stripPath := root
		if len(stripPath) > 1 && stripPath[len(stripPath)-1] == '/' {
			stripPath = stripPath[:len(stripPath)-1]
		}
		mux.Handle(root, http.StripPrefix(stripPath, handler))
		slog.Info("registered service", "root", root)
	}

	// health and root endpoints
	mux.HandleFunc("GET /health", handleHealth)

	srv := &http.Server{
		Addr:         m.listenAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("starting mockgrid HTTP server", "address", m.listenAddr)
	if err := srv.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			slog.Info("mockgrid server shutdown")
			return nil
		}
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// handleHealth returns a simple health check response.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy"}`))
}
