package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/model"
	"github.com/user/deer-flow-go/internal/store"
)

func setupThreadsTest(t *testing.T) (*ThreadsHandler, string) {
	tmpDir, err := os.MkdirTemp("", "threads-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return NewThreadsHandler(store.NewThreadStore(tmpDir)), tmpDir
}

func TestThreadsHandler_Create(t *testing.T) {
	handler, tmpDir := setupThreadsTest(t)
	defer os.RemoveAll(tmpDir)

	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads", nil)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["thread_id"] == "" {
		t.Error("expected thread_id in response")
	}
}

func TestThreadsHandler_Get(t *testing.T) {
	handler, tmpDir := setupThreadsTest(t)
	defer os.RemoveAll(tmpDir)

	thread, err := handler.threadStore.Create()
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/langgraph/threads/{id}", nil)
	req.SetPathValue("id", thread.ThreadID)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["thread_id"] != thread.ThreadID {
		t.Errorf("expected thread_id=%s, got %s", thread.ThreadID, resp["thread_id"])
	}
}

func TestThreadsHandler_Get_NotFound(t *testing.T) {
	handler, tmpDir := setupThreadsTest(t)
	defer os.RemoveAll(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/langgraph/threads/{id}", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestThreadsHandler_Delete(t *testing.T) {
	handler, tmpDir := setupThreadsTest(t)
	defer os.RemoveAll(tmpDir)

	thread, err := handler.threadStore.Create()
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/langgraph/threads/{id}", nil)
	req.SetPathValue("id", thread.ThreadID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	_, err = handler.threadStore.Get(thread.ThreadID)
	if err != store.ErrThreadNotFound {
		t.Errorf("expected ErrThreadNotFound after delete, got %v", err)
	}
}

func TestThreadsHandler_Delete_NotFound(t *testing.T) {
	handler, tmpDir := setupThreadsTest(t)
	defer os.RemoveAll(tmpDir)

	req := httptest.NewRequest(http.MethodDelete, "/api/langgraph/threads/{id}", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRouter_ThreadEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "router-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Storage: config.StorageConfig{DataDir: tmpDir},
	}
	router := NewRouter(cfg, "config.yaml")

	mux := http.NewServeMux()
	router.Register(mux)

	createReq := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads", nil)
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var thread map[string]any
	json.Unmarshal(createRec.Body.Bytes(), &thread)
	threadID := thread["thread_id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/api/langgraph/threads/"+threadID, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/langgraph/threads/"+threadID, nil)
	delRec := httptest.NewRecorder()
	mux.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, delRec.Code)
	}
}

func TestRunsHandler_Create(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runs-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	threadStore := store.NewThreadStore(tmpDir)
	runStore := store.NewRunStore(tmpDir)
	handler := NewRunsHandler(runStore)

	thread, err := threadStore.Create()
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	body := bytes.NewBufferString(`{"assistant_id": "agent", "input": {"message": "hello"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/{id}/runs", body)
	req.SetPathValue("id", thread.ThreadID)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["run_id"] == "" {
		t.Error("expected run_id in response")
	}
	if resp["thread_id"] != thread.ThreadID {
		t.Errorf("expected thread_id=%s, got %s", thread.ThreadID, resp["thread_id"])
	}
}

func TestRunsHandler_Get(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runs-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	threadStore := store.NewThreadStore(tmpDir)
	runStore := store.NewRunStore(tmpDir)
	handler := NewRunsHandler(runStore)

	thread, _ := threadStore.Create()
	run, _ := runStore.Create(thread.ThreadID, "agent", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/langgraph/threads/{id}/runs/{run_id}", nil)
	req.SetPathValue("id", thread.ThreadID)
	req.SetPathValue("run_id", run.RunID)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["run_id"] != run.RunID {
		t.Errorf("expected run_id=%s, got %s", run.RunID, resp["run_id"])
	}
}

func TestRunsHandler_Get_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runs-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	threadStore := store.NewThreadStore(tmpDir)
	runStore := store.NewRunStore(tmpDir)
	handler := NewRunsHandler(runStore)

	thread, _ := threadStore.Create()

	req := httptest.NewRequest(http.MethodGet, "/api/langgraph/threads/{id}/runs/{run_id}", nil)
	req.SetPathValue("id", thread.ThreadID)
	req.SetPathValue("run_id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRunsJoinHandler_Join(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runsjoin-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	threadStore := store.NewThreadStore(tmpDir)
	runStore := store.NewRunStore(tmpDir)
	handler := NewRunsJoinHandler(runStore)

	thread, _ := threadStore.Create()
	run, _ := runStore.Create(thread.ThreadID, "agent", nil)

	go func() {
		time.Sleep(100 * time.Millisecond)
		runStore.Update(thread.ThreadID, run.RunID, func(r *model.Run) {
			r.Status = model.RunStatusSuccess
			r.Output = map[string]any{"result": "done"}
		})
	}()

	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/{id}/runs/{run_id}/join", nil)
	req.SetPathValue("id", thread.ThreadID)
	req.SetPathValue("run_id", run.RunID)
	rec := httptest.NewRecorder()

	handler.Join(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	runResp := resp["run"].(map[string]any)
	if runResp["status"] != "success" {
		t.Errorf("expected status=success, got %s", runResp["status"])
	}
}
