package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mustur/mockgrid/app/api"
	"github.com/mustur/mockgrid/app/api/middleware"
	"github.com/mustur/mockgrid/internal/testutil"
)

// buildTestMux constructs a ServeMux from services the same way MockGrid.Start does.
// This allows testing the routing logic without starting a server.
func buildTestMux(services ...api.Service) *http.ServeMux {
	mux := http.NewServeMux()
	for _, svc := range services {
		root := svc.GetRoot()
		handler := svc.Chain()(svc.GetMux())
		// StripPrefix needs the path without trailing slash to avoid redirect issues
		stripPath := root
		if len(stripPath) > 1 && stripPath[len(stripPath)-1] == '/' {
			stripPath = stripPath[:len(stripPath)-1]
		}
		mux.Handle(root, http.StripPrefix(stripPath, handler))
	}
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	return mux
}

func TestMockGrid_NoServicesRegistered(t *testing.T) {
	mg := api.New(":0")
	err := mg.Start()
	if err == nil {
		t.Error("expected error when no services registered")
	}
	if err.Error() != "no services registered" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockGrid_HealthEndpoint(t *testing.T) {
	svc := testutil.NewMockService("/api/")
	svc.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := buildTestMux(svc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestMockGrid_ServiceRouting(t *testing.T) {
	svc1Called := false
	svc2Called := false

	svc1 := testutil.NewMockService("/v1/")
	svc1.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		svc1Called = true
		w.WriteHeader(http.StatusOK)
	})

	svc2 := testutil.NewMockService("/v2/")
	svc2.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		svc2Called = true
		w.WriteHeader(http.StatusOK)
	})

	mux := buildTestMux(svc1, svc2)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Request to /v1/test
	resp, err := http.Get(srv.URL + "/v1/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if !svc1Called {
		t.Error("svc1 should have been called")
	}
	if svc2Called {
		t.Error("svc2 should not have been called")
	}

	// Reset and request to /v2/test
	svc1Called = false
	resp, err = http.Get(srv.URL + "/v2/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if svc1Called {
		t.Error("svc1 should not have been called")
	}
	if !svc2Called {
		t.Error("svc2 should have been called")
	}
}

func TestMockGrid_ServiceMiddlewareApplied(t *testing.T) {
	recorder := testutil.NewRecordingMiddleware()

	svc := testutil.NewMockService("/api/")
	svc.WithChain(middleware.Chain(
		recorder.Middleware("auth"),
		recorder.Middleware("logging"),
	))
	svc.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := buildTestMux(svc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	records := recorder.Records()
	if len(records) != 4 {
		t.Fatalf("expected 4 middleware calls, got %d: %+v", len(records), records)
	}

	// Check order: auth-before, logging-before, logging-after, auth-after
	expected := []struct{ name, phase string }{
		{"auth", "before"},
		{"logging", "before"},
		{"logging", "after"},
		{"auth", "after"},
	}
	for i, e := range expected {
		if records[i].Name != e.name || records[i].Phase != e.phase {
			t.Errorf("position %d: expected %s-%s, got %s-%s",
				i, e.name, e.phase, records[i].Name, records[i].Phase)
		}
	}
}

func TestMockGrid_NotFound(t *testing.T) {
	svc := testutil.NewMockService("/api/")
	svc.HandleFunc("GET /exists", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := buildTestMux(svc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/doesnotexist")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestMockGrid_MethodNotAllowed(t *testing.T) {
	svc := testutil.NewMockService("/api/")
	svc.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := buildTestMux(svc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// GET instead of POST
	resp, err := http.Get(srv.URL + "/api/submit")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Go's ServeMux returns 405 for method mismatch on explicit method patterns
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestMockGrid_StripPrefix(t *testing.T) {
	var receivedPath string

	svc := testutil.NewMockService("/api/v1/")
	svc.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mux := buildTestMux(svc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/users")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Path should be stripped of the service root
	if receivedPath != "/users" {
		t.Errorf("expected path '/users', got %q", receivedPath)
	}
}

func TestMockGrid_MultipleRoutes(t *testing.T) {
	getHandlerCalled := false
	postHandlerCalled := false

	svc := testutil.NewMockService("/api/")
	svc.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		getHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	svc.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		postHandlerCalled = true
		w.WriteHeader(http.StatusCreated)
	})

	mux := buildTestMux(svc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// GET request
	resp, _ := http.Get(srv.URL + "/api/items")
	resp.Body.Close()
	if !getHandlerCalled {
		t.Error("GET handler should have been called")
	}
	if postHandlerCalled {
		t.Error("POST handler should not have been called")
	}

	// POST request
	getHandlerCalled = false
	resp, _ = http.Post(srv.URL+"/api/items", "application/json", nil)
	resp.Body.Close()
	if getHandlerCalled {
		t.Error("GET handler should not have been called")
	}
	if !postHandlerCalled {
		t.Error("POST handler should have been called")
	}
}
