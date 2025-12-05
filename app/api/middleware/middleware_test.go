package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mustur/mockgrid/app/api/middleware"
)

// TestChain_ExecutionOrder verifies that middlewares execute in order on request
// and reverse order on response.
func TestChain_ExecutionOrder(t *testing.T) {
	var order []string

	mw := func(name string) middleware.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chain := middleware.Chain(mw("A"), mw("B"), mw("C"))
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	expected := []string{"A-before", "B-before", "C-before", "handler", "C-after", "B-after", "A-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// TestChain_EmptyChain verifies that an empty chain passes through to handler.
func TestChain_EmptyChain(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	chain := middleware.Chain()
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// TestChain_SingleMiddleware verifies a single middleware works correctly.
func TestChain_SingleMiddleware(t *testing.T) {
	headerSet := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			headerSet = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chain := middleware.Chain(mw)
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if !headerSet {
		t.Error("middleware was not called")
	}
	if rec.Header().Get("X-Test") != "middleware" {
		t.Error("header not set by middleware")
	}
}

// TestChain_MiddlewareCanAbort verifies middleware can stop the chain.
func TestChain_MiddlewareCanAbort(t *testing.T) {
	handlerCalled := false

	abortMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			// Do not call next.ServeHTTP
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	chain := middleware.Chain(abortMw)
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("handler should not have been called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// TestBallAndChain_BallExecutesLast verifies ball middleware wraps entire chain.
func TestBallAndChain_BallExecutesLast(t *testing.T) {
	var order []string

	mw := func(name string) middleware.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chain := middleware.BallAndChain(mw("ball"), mw("A"), mw("B"))
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	// Ball wraps the entire chain, so it goes first-before and last-after
	expected := []string{"ball-before", "A-before", "B-before", "handler", "B-after", "A-after", "ball-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// TestBallAndChain_EmptyChain verifies ball with empty chain still works.
func TestBallAndChain_EmptyChain(t *testing.T) {
	var order []string

	ball := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "ball-before")
			next.ServeHTTP(w, r)
			order = append(order, "ball-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chain := middleware.BallAndChain(ball)
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	expected := []string{"ball-before", "handler", "ball-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// BenchmarkChain_SmallChain benchmarks a chain of 3 middlewares.
func BenchmarkChain_SmallChain(b *testing.B) {
	noop := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	chain := middleware.Chain(noop, noop, noop)
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
}

// BenchmarkChain_LargeChain benchmarks a chain of 10 middlewares.
func BenchmarkChain_LargeChain(b *testing.B) {
	noop := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mws := make([]middleware.Middleware, 10)
	for i := range mws {
		mws[i] = noop
	}
	chain := middleware.Chain(mws...)
	wrapped := chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
}
