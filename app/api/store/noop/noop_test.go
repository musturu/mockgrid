package noop_test

import (
	"testing"

	"github.com/mustur/mockgrid/app/api/store"
	"github.com/mustur/mockgrid/app/api/store/noop"
)

// Note: noop store doesn't actually persist, so it has different behavior
// from other stores. We test the noop-specific behavior here.

func TestNoop_Save_ReturnsNil(t *testing.T) {
	s := noop.New()
	defer s.Close()

	msg := &store.Message{MsgID: "test", FromEmail: "a@b.com"}
	if err := s.SaveMSG(msg); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestNoop_Get_ReturnsEmpty(t *testing.T) {
	s := noop.New()
	defer s.Close()

	// Save something
	s.SaveMSG(&store.Message{MsgID: "test"})

	// Get returns empty (noop doesn't store)
	got, err := s.GetMSG(store.GetQuery{ID: "test"})
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}

func TestNoop_Get_WithStatus_ReturnsEmpty(t *testing.T) {
	s := noop.New()
	defer s.Close()

	got, err := s.GetMSG(store.GetQuery{Status: store.StatusDelivered})
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}

func TestNoop_Get_All_ReturnsEmpty(t *testing.T) {
	s := noop.New()
	defer s.Close()

	got, err := s.GetMSG(store.GetQuery{})
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}

func TestNoop_Close_ReturnsNil(t *testing.T) {
	s := noop.New()
	if err := s.Close(); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	// Idempotent
	if err := s.Close(); err != nil {
		t.Errorf("second close: expected nil error, got: %v", err)
	}
}

func TestNoop_MultipleSaves_NoError(t *testing.T) {
	s := noop.New()
	defer s.Close()

	for i := 0; i < 100; i++ {
		msg := &store.Message{MsgID: "msg-" + string(rune('a'+i%26))}
		if err := s.SaveMSG(msg); err != nil {
			t.Fatalf("Save %d failed: %v", i, err)
		}
	}
}
