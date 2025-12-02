package router

import (
	"encoding/base64"
	"log/slog"
	"net/http"
)

func OpenTracker(w http.ResponseWriter, r *http.Request) {

	qry := r.URL.Query()
	slog.Info("email open tracked", "query", qry)

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	pixel, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
	if err != nil {
		slog.Error("failed to decode tracking pixel", "err", err)
		pixel = []byte{}
	}
	_, _ = w.Write(pixel)
}

// NewMux builds a ServeMux from already-constructed handlers. This keeps router
// // purely a registrar and avoids importing application logic.
// func NewMux(sendHandler http.Handler) *http.ServeMux {
// 	mux := http.NewServeMux()
//
// 	mux.Handle("/v3/mail/send", sendHandler)
//
// 	// tracking pixel endpoint (GET) used to record email opens
// 	mux.Handle("/v3/mail/track/open", http.MethodGet, TrackOpenHandler())
//
// 	mux.Handle("/", (http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
// 		w.Header().Set("Content-Type", "application/json")
// 		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Email server is running"})
// 	})))
//
// 	mux.Handle("/health", methodHandler(http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
// 		w.Header().Set("Content-Type", "application/json")
// 		_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
// 	})))
//
// 	return mux
// }
