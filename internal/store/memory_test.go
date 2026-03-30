package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryStore_Load_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	store := NewMemoryStore(storagePath)
	mem, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if mem == nil {
		t.Fatal("Load() returned nil")
	}

	if mem.WorkContext != "" {
		t.Errorf("WorkContext = %q, want empty", mem.WorkContext)
	}
}

func TestMemoryStore_Load_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	existingMem := &UserMemory{
		WorkContext:     "test context",
		PersonalContext: "personal info",
		Facts: []Fact{
			{ID: "1", Content: "fact 1", Category: "work", Confidence: 0.9},
		},
	}

	data, err := json.MarshalIndent(existingMem, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(storagePath, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	store := NewMemoryStore(storagePath)
	mem, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if mem.WorkContext != "test context" {
		t.Errorf("WorkContext = %q, want %q", mem.WorkContext, "test context")
	}
	if mem.PersonalContext != "personal info" {
		t.Errorf("PersonalContext = %q, want %q", mem.PersonalContext, "personal info")
	}
	if len(mem.Facts) != 1 {
		t.Errorf("len(Facts) = %d, want 1", len(mem.Facts))
	}
	if mem.Facts[0].ID != "1" {
		t.Errorf("Facts[0].ID = %q, want %q", mem.Facts[0].ID, "1")
	}
}

func TestMemoryStore_Load_Cached(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	existingMem := &UserMemory{
		WorkContext: "cached context",
	}
	data, _ := json.MarshalIndent(existingMem, "", "  ")
	os.WriteFile(storagePath, data, 0o644)

	store := NewMemoryStore(storagePath)

	mem1, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	newData := &UserMemory{WorkContext: "modified"}
	data2, _ := json.MarshalIndent(newData, "", "  ")
	os.WriteFile(storagePath, data2, 0o644)

	mem2, err := store.Load()
	if err != nil {
		t.Fatalf("Load() second call error = %v", err)
	}

	if mem2.WorkContext != "cached context" {
		t.Errorf("Load() should return cached value, got %q, want %q", mem2.WorkContext, "cached context")
	}

	if mem1 != mem2 {
		t.Error("Load() should return same pointer for cached value")
	}
}

func TestMemoryStore_Save(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "subdir", "memory.json")

	store := NewMemoryStore(storagePath)

	mem := &UserMemory{
		WorkContext: "saved context",
		Facts: []Fact{
			{ID: "fact-1", Content: "test fact", Confidence: 0.85},
		},
	}

	if err := store.Save(mem); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}

	data, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	var loaded UserMemory
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if loaded.WorkContext != "saved context" {
		t.Errorf("loaded WorkContext = %q, want %q", loaded.WorkContext, "saved context")
	}
	if len(loaded.Facts) != 1 {
		t.Errorf("len(loaded.Facts) = %d, want 1", len(loaded.Facts))
	}
}

func TestMemoryStore_Save_CacheUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	store := NewMemoryStore(storagePath)

	mem := &UserMemory{
		WorkContext: "updated context",
	}

	if err := store.Save(mem); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.WorkContext != "updated context" {
		t.Errorf("Load() after Save() = %q, want %q", loaded.WorkContext, "updated context")
	}
}

func TestMemoryStore_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	initialMem := &UserMemory{WorkContext: "initial"}
	data, _ := json.MarshalIndent(initialMem, "", "  ")
	os.WriteFile(storagePath, data, 0o644)

	store := NewMemoryStore(storagePath)

	mem1, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if mem1.WorkContext != "initial" {
		t.Errorf("WorkContext = %q, want %q", mem1.WorkContext, "initial")
	}

	modifiedMem := &UserMemory{WorkContext: "modified"}
	data2, _ := json.MarshalIndent(modifiedMem, "", "  ")
	os.WriteFile(storagePath, data2, 0o644)

	mem2, err := store.Reload()
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	if mem2.WorkContext != "modified" {
		t.Errorf("Reload() WorkContext = %q, want %q", mem2.WorkContext, "modified")
	}
}

func TestMemoryStore_Save_Atomic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "memory.json")

	store := NewMemoryStore(storagePath)

	mem := &UserMemory{WorkContext: "atomic test"}
	if err := store.Save(mem); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	tmpPath := storagePath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("tmp file should not exist after Save(), tmpPath=%s", tmpPath)
	}
}

func TestFact_AllFields(t *testing.T) {
	fact := Fact{
		ID:         "test-id",
		Content:    "test content",
		Category:   "personal",
		Confidence: 0.95,
		CreatedAt:  "2024-01-01T00:00:00Z",
		Source:     "conversation",
	}

	if fact.ID != "test-id" {
		t.Errorf("ID = %q, want %q", fact.ID, "test-id")
	}
	if fact.Content != "test content" {
		t.Errorf("Content = %q, want %q", fact.Content, "test content")
	}
	if fact.Category != "personal" {
		t.Errorf("Category = %q, want %q", fact.Category, "personal")
	}
	if fact.Confidence != 0.95 {
		t.Errorf("Confidence = %f, want %f", fact.Confidence, 0.95)
	}
	if fact.CreatedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", fact.CreatedAt, "2024-01-01T00:00:00Z")
	}
	if fact.Source != "conversation" {
		t.Errorf("Source = %q, want %q", fact.Source, "conversation")
	}
}
