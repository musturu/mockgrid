package filesystem_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mustur/mockgrid/app/api/store"
	"github.com/mustur/mockgrid/app/api/store/filesystem"
	"github.com/mustur/mockgrid/internal/testutil"
)

func TestFilesystem_Contract(t *testing.T) {
	testutil.RunStoreContractTests(t, "filesystem", func(t *testing.T) store.MessageStore {
		dir := t.TempDir()
		s, err := filesystem.New(dir)
		if err != nil {
			t.Fatalf("failed to create filesystem store: %v", err)
		}
		return s
	})
}

func TestFilesystem_New_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "subdir", "nested")
	s, err := filesystem.New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer s.Close()

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}

	// Should be able to save
	msg := &store.Message{MsgID: "test", FromEmail: "a@b.com", ToEmail: "b@c.com", Timestamp: 1}
	if err := s.SaveMSG(msg); err != nil {
		t.Errorf("Save failed: %v", err)
	}
}

func TestFilesystem_Save_RequiresID(t *testing.T) {
	dir := t.TempDir()
	s, err := filesystem.New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer s.Close()

	msg := &store.Message{FromEmail: "a@b.com"} // no MsgID
	if err := s.SaveMSG(msg); err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestFilesystem_Save_SanitizesFilename(t *testing.T) {
	dir := t.TempDir()
	s, err := filesystem.New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer s.Close()

	// ID with path separators should be sanitized
	msg := &store.Message{MsgID: "test/id", FromEmail: "a@b.com", ToEmail: "b@c.com", Timestamp: 1}
	if err := s.SaveMSG(msg); err != nil {
		t.Errorf("Save failed: %v", err)
	}
}

func TestFilesystem_Get_NotFound_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	s, err := filesystem.New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer s.Close()

	_, err = s.GetMSG(store.GetQuery{ID: "nonexistent"})
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestFilesystem_Close_Idempotent(t *testing.T) {
	dir := t.TempDir()
	s, err := filesystem.New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Errorf("first Close failed: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}

func BenchmarkFilesystem_Save(b *testing.B) {
	dir := b.TempDir()
	s, err := filesystem.New(dir)
	if err != nil {
		b.Fatalf("New failed: %v", err)
	}
	defer s.Close()

	msg := &store.Message{
		MsgID:     "bench-msg",
		FromEmail: "a@b.com",
		ToEmail:   "c@d.com",
		Subject:   "Benchmark",
		HTMLBody:  "<p>Test</p>",
		Status:    store.StatusProcessed,
		Timestamp: 1700000000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.MsgID = "bench-" + string(rune('a'+i%26))
		if err := s.SaveMSG(msg); err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkFilesystem_Get(b *testing.B) {
	dir := b.TempDir()
	s, err := filesystem.New(dir)
	if err != nil {
		b.Fatalf("New failed: %v", err)
	}
	defer s.Close()

	// Create test message
	msg := &store.Message{
		MsgID:     "bench-get",
		FromEmail: "a@b.com",
		ToEmail:   "c@d.com",
		Subject:   "Benchmark",
		Status:    store.StatusProcessed,
		Timestamp: 1700000000,
	}
	if err := s.SaveMSG(msg); err != nil {
		b.Fatalf("Save failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.GetMSG(store.GetQuery{ID: "bench-get"}); err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}
