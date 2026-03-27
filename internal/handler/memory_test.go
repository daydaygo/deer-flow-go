package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/deer-flow-go/internal/store"
)

func TestMemoryHandler_Get(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	initialMem := &store.UserMemory{
		WorkContext:     "test work context",
		PersonalContext: "test personal context",
		Facts: []store.Fact{
			{ID: "fact-1", Content: "test fact", Category: "work"},
		},
	}
	data, _ := json.MarshalIndent(initialMem, "", "  ")
	os.WriteFile(storagePath, data, 0644)

	memoryStore := store.NewMemoryStore(storagePath)
	handler := NewMemoryHandler(memoryStore)

	req := httptest.NewRequest(http.MethodGet, "/api/memory", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var response store.UserMemory
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.WorkContext != "test work context" {
		t.Errorf("WorkContext = %q, want %q", response.WorkContext, "test work context")
	}
	if response.PersonalContext != "test personal context" {
		t.Errorf("PersonalContext = %q, want %q", response.PersonalContext, "test personal context")
	}
	if len(response.Facts) != 1 {
		t.Errorf("len(Facts) = %d, want 1", len(response.Facts))
	}
}

func TestMemoryHandler_Get_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	memoryStore := store.NewMemoryStore(storagePath)
	handler := NewMemoryHandler(memoryStore)

	req := httptest.NewRequest(http.MethodGet, "/api/memory", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response store.UserMemory
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.WorkContext != "" {
		t.Errorf("WorkContext = %q, want empty", response.WorkContext)
	}
}

func TestMemoryHandler_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	initialMem := &store.UserMemory{
		WorkContext: "initial context",
	}
	data, _ := json.MarshalIndent(initialMem, "", "  ")
	os.WriteFile(storagePath, data, 0644)

	memoryStore := store.NewMemoryStore(storagePath)
	handler := NewMemoryHandler(memoryStore)

	req1 := httptest.NewRequest(http.MethodGet, "/api/memory", nil)
	rec1 := httptest.NewRecorder()
	handler.Get(rec1, req1)

	var response1 store.UserMemory
	json.Unmarshal(rec1.Body.Bytes(), &response1)
	if response1.WorkContext != "initial context" {
		t.Errorf("first Get() WorkContext = %q, want %q", response1.WorkContext, "initial context")
	}

	modifiedMem := &store.UserMemory{
		WorkContext: "modified context",
	}
	data2, _ := json.MarshalIndent(modifiedMem, "", "  ")
	os.WriteFile(storagePath, data2, 0644)

	req2 := httptest.NewRequest(http.MethodPost, "/api/memory/reload", nil)
	rec2 := httptest.NewRecorder()
	handler.Reload(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Reload() status = %d, want %d", rec2.Code, http.StatusOK)
	}

	var response2 store.UserMemory
	if err := json.Unmarshal(rec2.Body.Bytes(), &response2); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response2.WorkContext != "modified context" {
		t.Errorf("Reload() WorkContext = %q, want %q", response2.WorkContext, "modified context")
	}
}
