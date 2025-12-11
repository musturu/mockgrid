package sendmail

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"github.com/mustur/mockgrid/app/api/middleware"
	"github.com/mustur/mockgrid/app/api/objects"
	"github.com/mustur/mockgrid/app/api/store"
	"github.com/mustur/mockgrid/app/template"
)

// Config holds configuration for the SendMail service.
type Config struct {
	SMTPServer    string
	SMTPPort      int
	ListenAddr    string
	AttachmentDir string
	AuthKey       string
	SMTPUser      string
	SMTPPass      string
}

// Service implements the mail sending functionality.
type Service struct {
	smtpServer    string
	smtpPort      int
	listenAddr    string
	attachmentDir string
	authKey       string
	smtpUser      string
	smtpPass      string
	tpl           template.Templater
	store         store.MessageStore
}

// New creates a new SendMail service with the given configuration.
func New(cfg Config, tpl template.Templater, msgStore store.MessageStore) *Service {
	return &Service{
		smtpServer:    cfg.SMTPServer,
		smtpPort:      cfg.SMTPPort,
		listenAddr:    cfg.ListenAddr,
		attachmentDir: cfg.AttachmentDir,
		authKey:       cfg.AuthKey,
		smtpUser:      cfg.SMTPUser,
		smtpPass:      cfg.SMTPPass,
		tpl:           tpl,
		store:         msgStore,
	}
}

// authMiddleware returns a middleware that checks for valid authorization.
func (s *Service) authMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := s.checkAuth(r); err != nil {
				slog.Warn("authorization failed", "err", err, "path", r.URL.Path)
				writeJSON(w, http.StatusUnauthorized, objects.GetErrorResponse(err.Error(), nil, nil))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// checkAuth validates the Authorization header against the configured key.
func (s *Service) checkAuth(r *http.Request) error {
	if s.authKey == "" {
		return nil
	}
	authHeader := r.Header.Get("Authorization")
	expected := "Bearer " + s.authKey
	if authHeader != expected {
		return fmt.Errorf("the provided authorization grant is invalid, expired, or revoked")
	}
	return nil
}

// handleSend processes POST /v3/mail/send requests.
func (s *Service) handleSend(w http.ResponseWriter, r *http.Request) {
	if !validateContentType(w, r, "application/json") {
		return
	}

	pr, err := decodePostRequest(r)
	if err != nil {
		slog.Error("failed to decode request body", "err", err)
		writeJSON(w, http.StatusBadRequest, objects.GetErrorResponse("Failed to decode request body: "+err.Error(), nil, nil))
		return
	}

	if err := s.renderTemplate(pr); err != nil {
		slog.Error("failed to render template", "err", err)
		writeJSON(w, http.StatusInternalServerError, objects.GetErrorResponse("Failed to render template: "+err.Error(), nil, nil))
		return
	}

	if code, errResp := pr.Validate(); code != http.StatusAccepted {
		slog.Warn("validation failed", "status", code)
		writeJSON(w, code, errResp)
		return
	}

	if code, errResp := s.sendMail(pr); code != http.StatusAccepted {
		slog.Error("failed to send email", "status", code)
		writeJSON(w, code, errResp)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Email sent successfully"}); err != nil {
		slog.Error("failed to encode success response", "err", err)
	}
}

// handleTrackOpen serves the tracking pixel and logs the open event.
func (s *Service) handleTrackOpen(w http.ResponseWriter, r *http.Request) {
	qry := r.URL.Query()
	slog.Info("email open tracked", "id", qry.Get("id"), "to", qry.Get("to"))

	pixel, err := base64.StdEncoding.DecodeString(trackingPixelB64)
	if err != nil {
		slog.Error("failed to decode tracking pixel", "err", err)
		pixel = []byte{}
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pixel)
}

// sendMail iterates over personalizations and sends an email for each.
func (s *Service) sendMail(pr *objects.PostRequest) (int, objects.ErrorResponse) {
	auth := s.smtpAuth()

	for _, p := range pr.Personalizations {
		e := s.buildEmail(pr, p)

		s.injectTrackingPixels(e, p)

		if code, errResp := s.attachFiles(e, pr.Attachments); code != http.StatusAccepted {
			return code, errResp
		}

		sendErr := e.Send(s.smtpAddr(), auth)
		status, reason := classifyDeliveryResult(sendErr)

		if err := s.saveMessages(pr, p, e, status, reason); err != nil {
			slog.Error("failed to save messages", "err", err)
		}

		if sendErr != nil {
			slog.Error("failed to send email", "err", sendErr)
			return http.StatusInternalServerError, objects.GetErrorResponse("Failed to send email: "+sendErr.Error(), nil, nil)
		}
	}
	return http.StatusAccepted, objects.GetErrorResponse("", nil, nil)
}

// buildEmail constructs an email.Email from the request and personalization.
func (s *Service) buildEmail(pr *objects.PostRequest, p objects.Personalization) *email.Email {
	e := email.NewEmail()
	e.From = formatAddress(pr.From.Name, pr.From.Email)

	for _, to := range p.To {
		e.To = append(e.To, formatAddress(to.Name, to.Email))
	}
	for _, cc := range p.Cc {
		e.Cc = append(e.Cc, formatAddress(cc.Name, cc.Email))
	}
	for _, bcc := range p.Bcc {
		e.Bcc = append(e.Bcc, formatAddress(bcc.Name, bcc.Email))
	}

	replacer := buildReplacer(p.Substitutions)
	e.Subject = s.resolveSubject(pr, p, replacer)

	for _, c := range pr.Content {
		if c.Type == "text/html" {
			e.HTML = []byte(replacer.Replace(c.Value))
		} else {
			e.Text = []byte(replacer.Replace(c.Value))
		}
	}

	return e
}

// injectTrackingPixels adds tracking pixels to the email HTML body.
func (s *Service) injectTrackingPixels(e *email.Email, p objects.Personalization) {
	base := s.trackingBaseURL()
	for idx, to := range p.To {
		ensureHTMLBody(e)
		id := generateTrackingID(idx)
		trackURL := buildTrackingURL(base, id, to.Email)
		pixel := fmt.Sprintf(`<img src="%s" alt="" width="1" height="1" style="display:none;"/>`, trackURL)
		injectPixel(e, pixel)
	}
}

// attachFiles decodes and attaches files to the email.
func (s *Service) attachFiles(e *email.Email, attachments []objects.Attachment) (int, objects.ErrorResponse) {
	for i, att := range attachments {
		dir, err := s.saveAttachment(att.Filename, att.Content)
		if err != nil {
			slog.Error("attachment error", "filename", att.Filename, "err", err)
			return http.StatusBadRequest, objects.GetErrorResponse(
				"The attachment content must be base64 encoded.",
				"attachments."+strconv.Itoa(i)+".content",
				"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.attachments.content",
			)
		}
		path := filepath.Join(dir, filepath.Base(att.Filename))
		if _, err := e.AttachFile(path); err != nil {
			slog.Error("failed to attach file", "filename", att.Filename, "err", err)
			return http.StatusInternalServerError, objects.GetErrorResponse("Failed to attach file: "+err.Error(), nil, nil)
		}
	}
	return http.StatusAccepted, objects.GetErrorResponse("", nil, nil)
}

// saveAttachment decodes base64 content and writes it to a temporary directory.
func (s *Service) saveAttachment(filename, b64Content string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64Content)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	safeName := filepath.Base(filename)
	dir := filepath.Join(s.attachmentDir, "attachment_"+strconv.FormatInt(time.Now().UnixNano(), 10))

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	path := filepath.Join(dir, safeName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return dir, nil
}

// renderTemplate applies template rendering if a templater is configured.
func (s *Service) renderTemplate(pr *objects.PostRequest) error {
	if s.tpl == nil {
		return nil
	}
	return template.RenderAndPopulateFromTemplate(pr, s.tpl)
}

// resolveSubject returns the subject from personalization or request, with substitutions applied.
func (s *Service) resolveSubject(pr *objects.PostRequest, p objects.Personalization, r *strings.Replacer) string {
	if p.Subject != "" {
		return r.Replace(p.Subject)
	}
	return r.Replace(pr.Subject)
}

// smtpAddr returns the SMTP server address in host:port format.
func (s *Service) smtpAddr() string {
	return s.smtpServer + ":" + strconv.Itoa(s.smtpPort)
}

// smtpAuth returns SMTP authentication if credentials are configured.
// Returns nil if no credentials are set (anonymous SMTP).
func (s *Service) smtpAuth() smtp.Auth {
	if s.smtpUser == "" || s.smtpPass == "" {
		return nil
	}
	return smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpServer)
}

// saveMessages persists message records for each recipient.
func (s *Service) saveMessages(pr *objects.PostRequest, p objects.Personalization, e *email.Email, status store.MessageStatus, reason string) error {
	now := time.Now().Unix()

	for _, to := range p.To {
		msgID, err := store.GenerateMessageID()
		if err != nil {
			return fmt.Errorf("generate message ID: %w", err)
		}

		msg := &store.Message{
			MsgID:         msgID,
			FromEmail:     pr.From.Email,
			ToEmail:       to.Email,
			Subject:       e.Subject,
			HTMLBody:      string(e.HTML),
			TextBody:      string(e.Text),
			Status:        status,
			Reason:        reason,
			Timestamp:     now,
			LastEventTime: now,
		}

		if err := s.store.SaveMSG(msg); err != nil {
			slog.Error("failed to save message", "err", err, "msg_id", msgID)
		}
	}

	return nil
}

// trackingBaseURL builds the base URL for tracking endpoints.
func (s *Service) trackingBaseURL() string {
	base := s.listenAddr
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + strings.ReplaceAll(base, "0.0.0.0", "localhost")
	}
	return strings.TrimRight(base, "/")
}

// --- Pure helper functions (stateless, reusable) ---

const trackingPixelB64 = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"

// classifyDeliveryResult determines the message status based on SMTP response.
func classifyDeliveryResult(err error) (store.MessageStatus, string) {
	if err == nil {
		return store.StatusDelivered, ""
	}

	errStr := err.Error()

	// Permanent failures (5xx) - hard bounce
	if strings.Contains(errStr, "550") ||
		strings.Contains(errStr, "551") ||
		strings.Contains(errStr, "552") ||
		strings.Contains(errStr, "553") ||
		strings.Contains(errStr, "554") {
		return store.StatusBounce, errStr
	}

	// Temporary failures (4xx) - soft bounce / blocked
	if strings.Contains(errStr, "421") ||
		strings.Contains(errStr, "450") ||
		strings.Contains(errStr, "451") ||
		strings.Contains(errStr, "452") {
		return store.StatusBlocked, errStr
	}

	// Connection errors - deferred
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "dial") {
		return store.StatusDeferred, errStr
	}

	// Default to bounce for unknown errors
	return store.StatusBounce, errStr
}

// writeJSON encodes a response as JSON and writes it to the response writer.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

// validateContentType checks the Content-Type header and writes an error if invalid.
func validateContentType(w http.ResponseWriter, r *http.Request, expected string) bool {
	ct := r.Header.Get("Content-Type")
	if ct != expected {
		slog.Warn("invalid content-type", "got", ct, "expected", expected)
		writeJSON(w, http.StatusUnsupportedMediaType, objects.GetErrorResponse(
			"Content-Type should be "+expected, nil, nil))
		return false
	}
	return true
}

// decodePostRequest decodes the JSON request body into a PostRequest.
func decodePostRequest(r *http.Request) (*objects.PostRequest, error) {
	pr := &objects.PostRequest{}
	if err := json.NewDecoder(r.Body).Decode(pr); err != nil {
		return nil, err
	}
	return pr, nil
}

// formatAddress formats an email address with optional name.
func formatAddress(name, email string) string {
	if name == "" {
		return email
	}
	return name + " <" + email + ">"
}

// buildReplacer creates a strings.Replacer from a substitution map.
func buildReplacer(subs map[string]string) *strings.Replacer {
	pairs := make([]string, 0, len(subs)*2)
	for k, v := range subs {
		pairs = append(pairs, k, v)
	}
	return strings.NewReplacer(pairs...)
}

// generateTrackingID returns a unique identifier for tracking.
func generateTrackingID(idx int) string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + strconv.Itoa(idx)
}

// buildTrackingURL constructs a full tracking URL with query parameters.
func buildTrackingURL(base, id, to string) string {
	vals := url.Values{}
	vals.Set("id", id)
	vals.Set("to", to)
	return base + "/v3/mail/track/open?" + vals.Encode()
}

// ensureHTMLBody wraps plain text in basic HTML if no HTML body exists.
func ensureHTMLBody(e *email.Email) {
	if len(e.HTML) == 0 {
		e.HTML = []byte("<html><body>" + string(e.Text) + "</body></html>")
	}
}

// injectPixel inserts the tracking pixel before </body> or appends it.
func injectPixel(e *email.Email, pixel string) {
	html := string(e.HTML)
	if strings.Contains(strings.ToLower(html), "</body>") {
		e.HTML = []byte(strings.Replace(html, "</body>", pixel+"</body>", 1))
	} else {
		e.HTML = []byte(html + pixel)
	}
}
