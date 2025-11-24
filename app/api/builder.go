package api

import (
	"fmt"

	"github.com/mustur/mockgrid/app/template"
)

// Builder-style construction so callers (eg. cmd) don't need to import app/api's
// config types. Use NewBuilder().WithX(...).Build().
type MockGridBuilder struct {
	tpl               template.Templater
	smtpServer        string
	smtpPort          int
	listenHost        string
	listenPort        int
	attachmentDir     string
	enableAttachments bool
	sendgridAuthKey   string
}

func NewBuilder() *MockGridBuilder { return &MockGridBuilder{} }

func (b *MockGridBuilder) WithTemplate(t template.Templater) *MockGridBuilder { b.tpl = t; return b }
func (b *MockGridBuilder) WithSMTP(server string, port int) *MockGridBuilder {
	b.smtpServer = server
	b.smtpPort = port
	return b
}
func (b *MockGridBuilder) WithListen(host string, port int) *MockGridBuilder {
	b.listenHost = host
	b.listenPort = port
	return b
}
func (b *MockGridBuilder) WithAttachments(dir string) *MockGridBuilder {
	if dir == "" {
		b.enableAttachments = false
		return b
	}
	b.attachmentDir = dir
	return b
}
func (b *MockGridBuilder) WithAuth(sendgridKey string) *MockGridBuilder {
	b.sendgridAuthKey = sendgridKey
	return b
}
func (b *MockGridBuilder) Build() *MockGrid {

	mg := &MockGrid{
		tpl:               b.tpl,
		enableAttachments: b.enableAttachments,
		smtpServer:        b.smtpServer,
		smtpPort:          b.smtpPort,
		SMTPServerURL:     fmt.Sprintf("%s:%d", b.smtpServer, b.smtpPort),
		listenAddr:        fmt.Sprintf("%s:%d", b.listenHost, b.listenPort),
		attachmentDir:     b.attachmentDir,
		auth:              &Auth{SendgridKey: b.sendgridAuthKey},
	}

	return mg
}
