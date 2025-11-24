package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	obj "github.com/mustur/mockgrid/app/api/objects"
)

func (m *MockGrid) Auth(r *http.Request) error {
	if m.auth == nil {
		return nil
	}
	authHeader := r.Header.Get("Authorization")
	expectedAuth := "Bearer " + m.auth.SendgridKey
	if authHeader != expectedAuth {
		return errors.New("the provided authorization grant is invalid, expired, or revoked")
	}
	return nil
}

// SendEmailHandler handles POST /v3/mail/send requests.
func (m *MockGrid) SendEmailHandler(w http.ResponseWriter, r *http.Request) {
	writeFunc := func(status int, errorResponse obj.ErrorResponse) {
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			slog.Error("Failed to encode response", "err", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}

	err := m.Auth(r)
	if err != nil {
		slog.Warn("Authorization failed", "err", err)
		writeFunc(http.StatusUnauthorized, obj.GetErrorResponse(err.Error(), nil, nil))
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		slog.Warn("Invalid Content-Type header", "got", contentType)
		writeFunc(http.StatusUnsupportedMediaType, obj.GetErrorResponse(
			"Content-Type should be application/json", nil, nil))
		return
	}

	postRequest := &obj.PostRequest{}
	if err := json.NewDecoder(r.Body).Decode(postRequest); err != nil {
		slog.Error("Failed to decode request body", "err", err)
		writeFunc(http.StatusBadRequest, obj.GetErrorResponse("Failed to decode request body: "+err.Error(), nil, nil))
		return
	}

	if err := m.RenderAndPopulateFromTemplate(postRequest); err != nil {
		slog.Error("Failed to render and populate template", "err", err)
		writeFunc(http.StatusInternalServerError, obj.GetErrorResponse("Failed to render template: "+err.Error(), nil, nil))
		return
	}

	statusCode, errorResponse := postRequest.Validate()
	if statusCode != http.StatusAccepted {
		slog.Warn("Validation failed", "status", statusCode, "error", errorResponse)
		writeFunc(statusCode, errorResponse)
		return
	}

	statusCode, errorResponse = m.SendMailWithSMTP(postRequest)
	if statusCode != http.StatusAccepted {
		slog.Error("Failed to send email", "status", statusCode, "error", errorResponse)
		writeFunc(statusCode, errorResponse)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Email sent successfully"}); err != nil {
		slog.Error("Failed to encode success response", "err", err)
		http.Error(w, "Failed to encode success response", http.StatusInternalServerError)
		return
	}
}
