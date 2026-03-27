package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/deer-flow-go/internal/model"
)

func TestRunStore_Create(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewRunStore(tmpDir)

	run, err := store.Create("thread-123", "agent", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	if run.RunID == "" {
		t.Error("expected run_id to be set")
	}
	if run.ThreadID != "thread-123" {
		t.Errorf("expected thread_id=thread-123, got %s", run.ThreadID)
	}
	if run.AssistantID != "agent" {
		t.Errorf("expected assistant_id=agent, got %s", run.AssistantID)
	}
	if run.Status != model.RunStatusPending {
		t.Errorf("expected status=pending, got %s", run.Status)
	}
}

func TestRunStore_Get(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewRunStore(tmpDir)

	created, err := store.Create("thread-123", "agent", nil)
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	got, err := store.Get("thread-123", created.RunID)
	if err != nil {
		t.Fatalf("failed to get run: %v", err)
	}

	if got.RunID != created.RunID {
		t.Errorf("expected run_id=%s, got %s", created.RunID, got.RunID)
	}

	_, err = store.Get("thread-123", "nonexistent")
	if err != ErrRunNotFound {
		t.Errorf("expected ErrRunNotFound, got %v", err)
	}
}

func TestRunStore_Update(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewRunStore(tmpDir)

	created, err := store.Create("thread-123", "agent", nil)
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	updated, err := store.Update("thread-123", created.RunID, func(r *model.Run) {
		r.Status = model.RunStatusSuccess
		r.Output = map[string]any{"result": "done"}
	})
	if err != nil {
		t.Fatalf("failed to update run: %v", err)
	}

	if updated.Status != model.RunStatusSuccess {
		t.Errorf("expected status=success, got %s", updated.Status)
	}
	if updated.Output["result"] != "done" {
		t.Errorf("expected output result=done, got %v", updated.Output["result"])
	}

	got, _ := store.Get("thread-123", created.RunID)
	if got.Status != model.RunStatusSuccess {
		t.Errorf("expected persisted status=success, got %s", got.Status)
	}
}

func TestRunStore_List(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewRunStore(tmpDir)

	_, err = store.Create("thread-123", "agent", nil)
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	_, err = store.Create("thread-123", "agent", nil)
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	runs, err := store.List("thread-123")
	if err != nil {
		t.Fatalf("failed to list runs: %v", err)
	}

	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}

	runs, err = store.List("nonexistent-thread")
	if err != nil {
		t.Fatalf("failed to list runs for nonexistent thread: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs for nonexistent thread, got %d", len(runs))
	}
}

func TestRunStore_ThreadIsolation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewRunStore(tmpDir)

	run1, _ := store.Create("thread-1", "agent", nil)
	run2, _ := store.Create("thread-2", "agent", nil)

	_, err = store.Get("thread-1", run2.RunID)
	if err != ErrRunNotFound {
		t.Errorf("expected ErrRunNotFound for cross-thread access, got %v", err)
	}

	got, err := store.Get("thread-1", run1.RunID)
	if err != nil {
		t.Fatalf("failed to get run from correct thread: %v", err)
	}
	if got.RunID != run1.RunID {
		t.Errorf("expected run_id=%s, got %s", run1.RunID, got.RunID)
	}
}

func TestRunStore_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store1 := NewRunStore(tmpDir)
	created, err := store1.Create("thread-123", "agent", map[string]any{"input": "test"})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	store2 := NewRunStore(tmpDir)
	got, err := store2.Get("thread-123", created.RunID)
	if err != nil {
		t.Fatalf("failed to get run from new store instance: %v", err)
	}

	if got.RunID != created.RunID {
		t.Errorf("expected run_id=%s, got %s", created.RunID, got.RunID)
	}
	if got.Input["input"] != "test" {
		t.Errorf("expected input test, got %v", got.Input["input"])
	}
}

func TestRunStore_FileStructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runstore-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewRunStore(tmpDir)
	run, err := store.Create("thread-123", "agent", nil)
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	runPath := filepath.Join(tmpDir, "threads", "thread-123", "runs", run.RunID+".json")
	if _, err := os.Stat(runPath); os.IsNotExist(err) {
		t.Errorf("expected run file at %s, but not found", runPath)
	}
}
