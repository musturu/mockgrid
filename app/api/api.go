package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/mustur/mockgrid/app/api/objects"
	"github.com/mustur/mockgrid/app/api/router"
	"github.com/mustur/mockgrid/app/template"
)

type MockGrid struct {
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

type Auth struct {
	SendgridKey string
}

// Start initializes or starts the email server.
func (m *MockGrid) Start() error {
	// wire in-package handler
	m.mux = router.NewMux(http.HandlerFunc(m.SendEmailHandler))

	if m.SMTPServerURL == "" {
		return errors.New("SMTP server is not configured, email server will not start")
	}
	if err := http.ListenAndServe(m.listenAddr, m.mux); err != nil {
		return errors.New("failed to start email server")
	}
	slog.Info("Mockgrid server started against SMTP server", "server", m.SMTPServerURL)
	slog.Info("Mockgrid server is listening", "address", m.listenAddr)

	return nil
}

func (m *MockGrid) RenderAndPopulateFromTemplate(pr *objects.PostRequest) error {
	if m.tpl != nil {
		return template.RenderAndPopulateFromTemplate(pr, m.tpl)
	}
	return nil
}
