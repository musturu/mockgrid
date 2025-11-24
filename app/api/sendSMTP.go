package api

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"github.com/mustur/mockgrid/app/api/objects"
)

func (m *MockGrid) SendMailWithSMTP(pr *objects.PostRequest) (int, objects.ErrorResponse) {
	// iterate over personalizations and send an email per personalization
	for _, personalizations := range pr.Personalizations {
		e := email.NewEmail()
		e.From = pr.From.Name + " <" + pr.From.Email + ">"

		for _, to := range personalizations.To {
			e.To = append(e.To, m.getEmailwithName(to))
		}
		for _, cc := range personalizations.Cc {
			e.Cc = append(e.Cc, m.getEmailwithName(cc))
		}
		for _, bcc := range personalizations.Bcc {
			e.Bcc = append(e.Bcc, m.getEmailwithName(bcc))
		}

		// prepare substitutions replacer
		replacements := make([]string, 0, len(personalizations.Substitutions)*2)
		for key, value := range personalizations.Substitutions {
			replacements = append(replacements, key, value)
		}
		replacer := strings.NewReplacer(replacements...)

		if personalizations.Subject != "" {
			e.Subject = replacer.Replace(personalizations.Subject)
		} else if pr.Subject != "" {
			e.Subject = replacer.Replace(pr.Subject)
		} else {
			return http.StatusBadRequest,
				objects.GetErrorResponse(
					"The subject is required. You can get around this requirement if you use a template with a subject defined or if every personalization has a subject defined.",
					"subject",
					"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.subject",
				)
		}

		for _, content := range pr.Content {
			if content.Type == "text/html" {
				e.HTML = []byte(replacer.Replace(content.Value))
			} else {
				e.Text = []byte(replacer.Replace(content.Value))
			}
		}

		// build base URL used for tracking pixels
		base := trackingBaseURL(m)

		// inject tracking pixels per recipient
		for idx := range personalizations.To {
			ensureHTMLBody(e, e.Text)
			id := generateTrackingID(idx)
			trackingURL := buildTrackingURL(base, id, personalizations.To[idx].Email)
			pixel := fmt.Sprintf(`<img src="%s" alt="" width="1" height="1" style="display:none;"/>`, trackingURL)
			injectTrackingPixel(e, pixel)
		}

		i := 0
		for _, attachment := range pr.Attachments {
			dirName := m.createAttachment(attachment.Filename, attachment.Content)
			if dirName == "" {
				slog.Error("Attachment content must be base64 encoded", "filename", attachment.Filename)
				return http.StatusBadRequest,
					objects.GetErrorResponse(
						"The attachment content must be base64 encoded.",
						"attachments."+strconv.Itoa(i)+".content",
						"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.attachments.content",
					)
			}
			_, err := e.AttachFile(filepath.Join(dirName, attachment.Filename))
			if err != nil {
				slog.Error("Failed to attach file", "filename", attachment.Filename, "err", err)
				return http.StatusInternalServerError, objects.GetErrorResponse("Failed to attach file: "+err.Error(), nil, nil)
			}
			i++
		}

		// send using configured smtp server/port without auth (legacy behaviour)
		if err := e.Send(m.smtpServer+":"+strconv.Itoa(m.smtpPort), nil); err != nil {
			slog.Error("Failed to send email without SMTP auth", "err", err)
			return http.StatusInternalServerError, objects.GetErrorResponse("Failed to send email: "+err.Error(), nil, nil)
		}
	}
	return http.StatusAccepted, objects.GetErrorResponse("", nil, nil)
}

func (m *MockGrid) getEmailwithName(t struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}) string {
	if t.Name == "" {
		return t.Email
	}
	return t.Name + " <" + t.Email + ">"
}

// createAttachment writes a base64 attachment into the configured directory and returns the directory path.
func (m *MockGrid) createAttachment(fileName string, base64Content string) string {
	data, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		slog.Error("Failed to decode base64 attachment", "filename", fileName, "err", err)
		return ""
	}

	dirName := filepath.Join(m.attachmentDir, "attachment_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	if err := os.MkdirAll(dirName, 0o777); err != nil {
		slog.Error("Failed to create attachment directory", "dir", dirName, "err", err)
		return ""
	}
	filePath := filepath.Join(dirName, fileName)
	if err := os.WriteFile(filePath, data, 0o666); err != nil {
		slog.Error("Failed to write attachment file", "file", filePath, "err", err)
		return ""
	}
	return dirName
}

// trackingBaseURL returns the base URL to use for tracking endpoints.
func trackingBaseURL(m *MockGrid) string {

	base := m.listenAddr
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + strings.ReplaceAll(base, "0.0.0.0", "localhost")
	}
	return strings.TrimRight(base, "/")
}

// generateTrackingID returns a unique id for a tracking pixel.
func generateTrackingID(idx int) string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + strconv.Itoa(idx)
}

// buildTrackingURL constructs the full tracking URL with query parameters.
func buildTrackingURL(base, id, to string) string {
	vals := url.Values{}
	vals.Set("id", id)
	vals.Set("to", to)
	return base + "/v3/mail/track/open?" + vals.Encode()
}

// ensureHTMLBody makes sure the email has an HTML body; if not, it wraps the plain text.
func ensureHTMLBody(e *email.Email, text []byte) {
	if len(e.HTML) == 0 {
		safeText := string(text)
		e.HTML = []byte("<html><body>" + safeText + "</body></html>")
	}
}

// injectTrackingPixel inserts the provided pixel HTML before </body> if present, otherwise appends it.
func injectTrackingPixel(e *email.Email, pixel string) {
	htmlStr := string(e.HTML)
	if strings.Contains(strings.ToLower(htmlStr), "</body>") {
		e.HTML = []byte(strings.Replace(htmlStr, "</body>", pixel+"</body>", 1))
	} else {
		e.HTML = []byte(htmlStr + pixel)
	}
}
