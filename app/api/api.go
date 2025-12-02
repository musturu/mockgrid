package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mustur/mockgrid/app/api/objects"
	"github.com/mustur/mockgrid/app/template"
)

type MockGrid struct {
	Services          []Service
	tpl               template.Templater
	mux               *http.ServeMux
	enableAttachments bool
	smtpServer        string
	smtpPort          int
	SMTPServerURL     string
	listenAddr        string
	attachmentDir     string
	auth              *Auth
}

type Service interface {
	GetMux() *http.ServeMux
	GetRoot() string
	Chain() Middleware
}

type Auth struct {
	SendgridKey string
}

// Start initializes or starts the email server.
func (m *MockGrid) Start() error {

	rout := http.NewServeMux()

	for _, s := range m.Services {
		rout.Handle(s.GetRoot(), http.StripPrefix(s.GetRoot(), s.Chain()(s.GetMux())))
	}

	if m.SMTPServerURL == "" {
		return errors.New("SMTP server is not configured, email server will not start")
	}
	srv := &http.Server{
		Addr:         m.listenAddr,
		Handler:      m.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("Starting Mockgrid HTTP server", "address", m.listenAddr)
	if err := srv.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			// graceful shutdown occurred
			slog.Info("Mockgrid server shutdown")
			return nil
		}
		return fmt.Errorf("failed to start email server: %w", err)
	}

	slog.Info("Mockgrid server started against SMTP server", "server", m.SMTPServerURL)
	return nil
}

func (m *MockGrid) RenderAndPopulateFromTemplate(pr *objects.PostRequest) error {
	if m.tpl != nil {
		return template.RenderAndPopulateFromTemplate(pr, m.tpl)
	}
	return nil
}
