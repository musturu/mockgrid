package sendmail

import (
	"net/http"

	"github.com/mustur/mockgrid/app/api/middleware"
)

// GetMux returns the service's HTTP multiplexer.
func (s *Service) GetMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /send", s.handleSend)
	mux.HandleFunc("GET /track/open", s.handleTrackOpen)
	return mux
}

// GetRoot returns the root path prefix for this service.
func (s *Service) GetRoot() string {
	return "/v3/mail/"
}

// Chain returns the middleware chain for this service.
func (s *Service) Chain() middleware.Middleware {
	return middleware.Chain(
		s.authMiddleware(),
	)
}
