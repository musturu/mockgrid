package objects

import (
	"errors"
	"net/http"

	"gopkg.in/go-playground/validator.v9"
)

// EmailAddress represents an email address with optional name.
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Personalization represents a single personalization block in a SendGrid request.
type Personalization struct {
	DynamicTemplateData map[string]interface{} `json:"dynamic_template_data"`
	To                  []EmailAddress         `json:"to"`
	Cc                  []EmailAddress         `json:"cc"`
	Bcc                 []EmailAddress         `json:"bcc"`
	Substitutions       map[string]string      `json:"substitutions"`
	Subject             string                 `json:"subject"`
}

// Content represents an email content block.
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Content     string `json:"content"`
	Type        string `json:"type"`
	Filename    string `json:"filename"`
	Disposition string `json:"disposition"`
	ContentId   string `json:"content_id"`
}

// PostRequest represents the structure of the email request body in SendGrid format.
type PostRequest struct {
	Personalizations []Personalization `json:"personalizations" validate:"required"`
	From             EmailAddress      `json:"from" validate:"required"`
	ReplyTo          EmailAddress      `json:"reply_to"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
	Attachments      []Attachment      `json:"attachments"`
	TemplateID       string            `json:"template_id"`
}

// Validate validates the PostRequest fields and returns appropriate error responses.
func (p *PostRequest) Validate() (int, ErrorResponse) {
	validate := validator.New()
	if err := validate.Struct(p); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, verr := range validationErrors {
				switch verr.ActualTag() {
				case "required":
					switch verr.StructField() {
					case "Personalizations":
						return http.StatusBadRequest,
							GetErrorResponse(
								"The personalizations field is required and must have at least one personalization.",
								"personalizations",
								"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#-Personalizations-Errors",
							)
					case "Email":
						return http.StatusBadRequest,
							GetErrorResponse(
								"The from object must be provided for every email send. It is an object that requires the email parameter, but may also contain a name parameter.  e.g. {\"email\" : \"example@example.com\"}  or {\"email\" : \"example@example.com\", \"name\" : \"Example Recipient\"}.",
								"from.email",
								"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.from",
							)
					case "Content":
						return http.StatusBadRequest,
							GetErrorResponse(
								"Unless a valid template_id is provided, the content parameter is required. There must be at least one defined content block. We typically suggest both text/plain and text/html blocks are included, but only one block is required.",
								"content",
								"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.content",
							)
					}
				}
			}
		} else {
			return http.StatusBadRequest, GetErrorResponse("Validation failed: "+err.Error(), nil, nil)
		}
	}
	return http.StatusAccepted, GetErrorResponse("", nil, nil)
}
