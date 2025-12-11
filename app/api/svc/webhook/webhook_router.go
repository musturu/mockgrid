// Package webhook provides webhook routing.
package webhook

import (
	"net/http"

	"github.com/mustur/mockgrid/app/api/middleware"
)

// GetMux returns the service's HTTP multiplexer.
func (s *Service) GetMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", s.HandleCreateWebhook)
	mux.HandleFunc("GET /", s.HandleListWebhooks)
	mux.HandleFunc("GET /{id}", s.HandleGetWebhook)
	mux.HandleFunc("PUT /{id}", s.HandleUpdateWebhook)
	mux.HandleFunc("DELETE /{id}", s.HandleDeleteWebhook)
	mux.HandleFunc("POST /{id}/toggle", s.HandleToggleWebhook)
	return mux
}

// GetRoot returns the root path prefix for this service.
func (s *Service) GetRoot() string {
	return "/v3/webhooks/"
}

// Chain returns the middleware chain for this service.
func (s *Service) Chain() middleware.Middleware {
	return middleware.Chain(
		s.authMiddleware(),
	)
}

// authMiddleware returns an authentication middleware
func (s *Service) authMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// noop, TODO
			next.ServeHTTP(w, r)
		})
	}
}
