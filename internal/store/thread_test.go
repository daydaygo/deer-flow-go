package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThreadStore_Create(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewThreadStore(tmpDir)

	thread, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if thread == nil {
		t.Fatal("Create() returned nil thread")
	}

	if thread.ThreadID == "" {
		t.Error("ThreadID is empty, should be set")
	}

	if thread.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, should be set")
	}

	if thread.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero, should be set")
	}

	threadPath := filepath.Join(tmpDir, "threads", thread.ThreadID, "thread.json")
	if _, err := os.Stat(threadPath); os.IsNotExist(err) {
		t.Errorf("thread file not created at %s", threadPath)
	}
}

func TestThreadStore_Get(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewThreadStore(tmpDir)

	created, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	thread, err := store.Get(created.ThreadID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if thread.ThreadID != created.ThreadID {
		t.Errorf("ThreadID = %q, want %q", thread.ThreadID, created.ThreadID)
	}
}

func TestThreadStore_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewThreadStore(tmpDir)

	_, err := store.Get("nonexistent-id")
	if err == nil {
		t.Fatal("Get() expected error for nonexistent thread, got nil")
	}

	if err != ErrThreadNotFound {
		t.Errorf("Get() error = %v, want ErrThreadNotFound", err)
	}
}

func TestThreadStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewThreadStore(tmpDir)

	thread, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Delete(thread.ThreadID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = store.Get(thread.ThreadID)
	if err == nil {
		t.Fatal("Get() after Delete() should return error")
	}

	if err != ErrThreadNotFound {
		t.Errorf("Get() after Delete() error = %v, want ErrThreadNotFound", err)
	}
}

func TestThreadStore_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewThreadStore(tmpDir)

	err := store.Delete("nonexistent-id")
	if err == nil {
		t.Fatal("Delete() expected error for nonexistent thread, got nil")
	}

	if err != ErrThreadNotFound {
		t.Errorf("Delete() error = %v, want ErrThreadNotFound", err)
	}
}

func TestThreadStore_Create_Atomic(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewThreadStore(tmpDir)

	thread, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	threadPath := filepath.Join(tmpDir, "threads", thread.ThreadID, "thread.json")
	tmpPath := threadPath + ".tmp"

	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("tmp file should not exist after Create(), tmpPath=%s", tmpPath)
	}
}
