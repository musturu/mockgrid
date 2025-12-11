package testutil

import (
	"testing"

	"github.com/mustur/mockgrid/app/api/store"
)

// StoreFactory creates a new MessageStore instance for testing.
// The store should be empty and ready for use.
type StoreFactory func(t *testing.T) store.MessageStore

// RunStoreContractTests runs the standard interface contract tests against any
// MessageStore implementation.
func RunStoreContractTests(t *testing.T, name string, factory StoreFactory) {
	t.Run(name+"/Save_and_GetByID", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		msg := &store.Message{
			MsgID:     "test-123",
			FromEmail: "sender@example.com",
			ToEmail:   "recipient@example.com",
			Subject:   "Test Subject",
			Status:    store.StatusProcessed,
			Timestamp: 1700000000,
		}

		if err := s.SaveMSG(msg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		got, err := s.GetMSG(store.GetQuery{ID: "test-123"})
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 message, got %d", len(got))
		}
		if got[0].MsgID != "test-123" {
			t.Errorf("expected MsgID 'test-123', got %q", got[0].MsgID)
		}
		if got[0].FromEmail != "sender@example.com" {
			t.Errorf("expected FromEmail 'sender@example.com', got %q", got[0].FromEmail)
		}
		if got[0].Status != store.StatusProcessed {
			t.Errorf("expected Status 'processed', got %q", got[0].Status)
		}
	})

	t.Run(name+"/Get_NonExistent_ReturnsEmptyOrNotFound", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		got, err := s.GetMSG(store.GetQuery{ID: "does-not-exist"})
		// Either returns empty slice with no error, or ErrNotFound
		if err != nil && err != store.ErrNotFound {
			t.Fatalf("unexpected error: %v", err)
		}
		if err == nil && len(got) != 0 {
			t.Errorf("expected empty result, got %d messages", len(got))
		}
	})

	t.Run(name+"/Save_Upsert", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		msg := &store.Message{
			MsgID:     "upsert-1",
			FromEmail: "sender@example.com",
			ToEmail:   "recipient@example.com",
			Status:    store.StatusProcessed,
			Timestamp: 1700000000,
		}
		if err := s.SaveMSG(msg); err != nil {
			t.Fatalf("first Save failed: %v", err)
		}

		// Update status
		msg.Status = store.StatusDelivered
		msg.LastEventTime = 1700000001
		if err := s.SaveMSG(msg); err != nil {
			t.Fatalf("upsert Save failed: %v", err)
		}

		got, err := s.GetMSG(store.GetQuery{ID: "upsert-1"})
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 message, got %d", len(got))
		}
		if got[0].Status != store.StatusDelivered {
			t.Errorf("upsert failed: expected status 'delivered', got %q", got[0].Status)
		}
	})

	t.Run(name+"/Get_FilterByStatus", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		// Save messages with different statuses
		msgs := []*store.Message{
			{MsgID: "msg-1", FromEmail: "a@b.com", ToEmail: "c@d.com", Status: store.StatusProcessed, Timestamp: 1},
			{MsgID: "msg-2", FromEmail: "a@b.com", ToEmail: "c@d.com", Status: store.StatusDelivered, Timestamp: 2},
			{MsgID: "msg-3", FromEmail: "a@b.com", ToEmail: "c@d.com", Status: store.StatusDelivered, Timestamp: 3},
		}
		for _, m := range msgs {
			if err := s.SaveMSG(m); err != nil {
				t.Fatalf("Save failed: %v", err)
			}
		}

		got, err := s.GetMSG(store.GetQuery{Status: store.StatusDelivered})
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("expected 2 delivered messages, got %d", len(got))
		}
		for _, m := range got {
			if m.Status != store.StatusDelivered {
				t.Errorf("expected status 'delivered', got %q", m.Status)
			}
		}
	})

	t.Run(name+"/Get_WithLimit", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		// Save multiple messages
		for i := 0; i < 5; i++ {
			msg := &store.Message{
				MsgID:     "limit-" + string(rune('a'+i)),
				FromEmail: "a@b.com",
				ToEmail:   "c@d.com",
				Status:    store.StatusProcessed,
				Timestamp: int64(i),
			}
			if err := s.SaveMSG(msg); err != nil {
				t.Fatalf("Save failed: %v", err)
			}
		}

		got, err := s.GetMSG(store.GetQuery{Limit: 3})
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(got) > 3 {
			t.Errorf("expected at most 3 messages, got %d", len(got))
		}
	})

	t.Run(name+"/Get_All", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		// Save multiple messages
		for i := 0; i < 3; i++ {
			msg := &store.Message{
				MsgID:     "all-" + string(rune('a'+i)),
				FromEmail: "a@b.com",
				ToEmail:   "c@d.com",
				Status:    store.StatusProcessed,
				Timestamp: int64(i),
			}
			if err := s.SaveMSG(msg); err != nil {
				t.Fatalf("Save failed: %v", err)
			}
		}

		got, err := s.GetMSG(store.GetQuery{})
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("expected 3 messages, got %d", len(got))
		}
	})

	t.Run(name+"/Close_Idempotent", func(t *testing.T) {
		s := factory(t)

		if err := s.Close(); err != nil {
			t.Errorf("first Close failed: %v", err)
		}
		if err := s.Close(); err != nil {
			t.Errorf("second Close failed: %v", err)
		}
	})

	t.Run(name+"/Save_PreservesAllFields", func(t *testing.T) {
		s := factory(t)
		defer s.Close()

		msg := &store.Message{
			MsgID:         "full-msg",
			FromEmail:     "sender@example.com",
			ToEmail:       "recipient@example.com",
			Subject:       "Test Subject",
			HTMLBody:      "<html><body>Hello</body></html>",
			TextBody:      "Hello",
			Status:        store.StatusDelivered,
			SMTPResponse:  "250 OK",
			Reason:        "",
			Timestamp:     1700000000,
			LastEventTime: 1700000001,
			OpensCount:    5,
			ClicksCount:   2,
		}

		if err := s.SaveMSG(msg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		got, err := s.GetMSG(store.GetQuery{ID: "full-msg"})
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 message, got %d", len(got))
		}

		g := got[0]
		if g.Subject != msg.Subject {
			t.Errorf("Subject: expected %q, got %q", msg.Subject, g.Subject)
		}
		if g.HTMLBody != msg.HTMLBody {
			t.Errorf("HTMLBody: expected %q, got %q", msg.HTMLBody, g.HTMLBody)
		}
		if g.TextBody != msg.TextBody {
			t.Errorf("TextBody: expected %q, got %q", msg.TextBody, g.TextBody)
		}
		if g.SMTPResponse != msg.SMTPResponse {
			t.Errorf("SMTPResponse: expected %q, got %q", msg.SMTPResponse, g.SMTPResponse)
		}
		if g.Timestamp != msg.Timestamp {
			t.Errorf("Timestamp: expected %d, got %d", msg.Timestamp, g.Timestamp)
		}
		if g.LastEventTime != msg.LastEventTime {
			t.Errorf("LastEventTime: expected %d, got %d", msg.LastEventTime, g.LastEventTime)
		}
		if g.OpensCount != msg.OpensCount {
			t.Errorf("OpensCount: expected %d, got %d", msg.OpensCount, g.OpensCount)
		}
		if g.ClicksCount != msg.ClicksCount {
			t.Errorf("ClicksCount: expected %d, got %d", msg.ClicksCount, g.ClicksCount)
		}
	})
}
