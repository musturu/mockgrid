package testutil

import (
	"github.com/mustur/mockgrid/app/api/objects"
	"github.com/mustur/mockgrid/app/api/store"
)

// NewTestPostRequest creates a minimal valid PostRequest for testing.
func NewTestPostRequest() *objects.PostRequest {
	return &objects.PostRequest{
		From: objects.EmailAddress{
			Email: "sender@example.com",
			Name:  "Test Sender",
		},
		Personalizations: []objects.Personalization{
			{
				To: []objects.EmailAddress{
					{Email: "recipient@example.com", Name: "Test Recipient"},
				},
			},
		},
		Subject: "Test Subject",
		Content: []objects.Content{
			{Type: "text/plain", Value: "Test body"},
		},
	}
}

// PostRequestBuilder provides a fluent API for building test PostRequests.
type PostRequestBuilder struct {
	pr *objects.PostRequest
}

// NewPostRequestBuilder creates a new builder with defaults.
func NewPostRequestBuilder() *PostRequestBuilder {
	return &PostRequestBuilder{pr: NewTestPostRequest()}
}

// WithFrom sets the sender.
func (b *PostRequestBuilder) WithFrom(email, name string) *PostRequestBuilder {
	b.pr.From = objects.EmailAddress{Email: email, Name: name}
	return b
}

// WithTo sets a single recipient.
func (b *PostRequestBuilder) WithTo(email, name string) *PostRequestBuilder {
	if len(b.pr.Personalizations) == 0 {
		b.pr.Personalizations = []objects.Personalization{{}}
	}
	b.pr.Personalizations[0].To = []objects.EmailAddress{{Email: email, Name: name}}
	return b
}

// WithSubject sets the subject.
func (b *PostRequestBuilder) WithSubject(subject string) *PostRequestBuilder {
	b.pr.Subject = subject
	return b
}

// WithContent sets the content.
func (b *PostRequestBuilder) WithContent(contentType, value string) *PostRequestBuilder {
	b.pr.Content = []objects.Content{{Type: contentType, Value: value}}
	return b
}

// WithTemplateID sets the template ID.
func (b *PostRequestBuilder) WithTemplateID(id string) *PostRequestBuilder {
	b.pr.TemplateID = id
	return b
}

// Build returns the constructed PostRequest.
func (b *PostRequestBuilder) Build() *objects.PostRequest {
	return b.pr
}

// NewTestMessage creates a minimal Message for testing.
func NewTestMessage(id string) *store.Message {
	return &store.Message{
		MsgID:     id,
		FromEmail: "sender@example.com",
		ToEmail:   "recipient@example.com",
		Subject:   "Test Subject",
		Status:    store.StatusProcessed,
		Timestamp: 1700000000,
	}
}

// MessageBuilder provides a fluent API for building test Messages.
type MessageBuilder struct {
	msg *store.Message
}

// NewMessageBuilder creates a new builder with defaults.
func NewMessageBuilder(id string) *MessageBuilder {
	return &MessageBuilder{msg: NewTestMessage(id)}
}

// WithFrom sets the sender email.
func (b *MessageBuilder) WithFrom(email string) *MessageBuilder {
	b.msg.FromEmail = email
	return b
}

// WithTo sets the recipient email.
func (b *MessageBuilder) WithTo(email string) *MessageBuilder {
	b.msg.ToEmail = email
	return b
}

// WithSubject sets the subject.
func (b *MessageBuilder) WithSubject(subject string) *MessageBuilder {
	b.msg.Subject = subject
	return b
}

// WithStatus sets the delivery status.
func (b *MessageBuilder) WithStatus(status store.MessageStatus) *MessageBuilder {
	b.msg.Status = status
	return b
}

// WithHTMLBody sets the HTML body.
func (b *MessageBuilder) WithHTMLBody(html string) *MessageBuilder {
	b.msg.HTMLBody = html
	return b
}

// WithTextBody sets the text body.
func (b *MessageBuilder) WithTextBody(text string) *MessageBuilder {
	b.msg.TextBody = text
	return b
}

// WithTimestamp sets the timestamp.
func (b *MessageBuilder) WithTimestamp(ts int64) *MessageBuilder {
	b.msg.Timestamp = ts
	return b
}

// Build returns the constructed Message.
func (b *MessageBuilder) Build() *store.Message {
	return b.msg
}
