package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/deer-flow-go/internal/skills"
)

func setupSkillsTestEnv(t *testing.T) (*skills.Loader, string) {
	t.Helper()

	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")
	publicDir := filepath.Join(skillsDir, "public")
	os.MkdirAll(publicDir, 0755)

	skillDir := filepath.Join(publicDir, "test-skill")
	os.MkdirAll(skillDir, 0755)
	skillContent := `---
name: test-skill
description: Test skill for API
license: MIT
---
# Test Skill

Test content.
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)

	configPath := filepath.Join(tmpDir, "extensions_config.json")
	os.WriteFile(configPath, []byte("{}"), 0644)

	loader := skills.NewLoader(skillsDir, configPath)
	if err := loader.Load(); err != nil {
		t.Fatalf("Failed to load skills: %v", err)
	}

	return loader, configPath
}

func TestSkillsHandler_List(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var response SkillsListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(response.Skills) != 1 {
		t.Errorf("len(Skills) = %d, want 1", len(response.Skills))
	}

	if response.Skills[0].Name != "test-skill" {
		t.Errorf("Skills[0].Name = %q, want %q", response.Skills[0].Name, "test-skill")
	}
}

func TestSkillsHandler_Get_Found(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	req := httptest.NewRequest(http.MethodGet, "/api/skills/test-skill", nil)
	req.SetPathValue("name", "test-skill")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var skill skills.Skill
	if err := json.Unmarshal(rec.Body.Bytes(), &skill); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
	}
	if skill.Description != "Test skill for API" {
		t.Errorf("Description = %q, want %q", skill.Description, "Test skill for API")
	}
	if skill.License != "MIT" {
		t.Errorf("License = %q, want %q", skill.License, "MIT")
	}
}

func TestSkillsHandler_Get_NotFound(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	req := httptest.NewRequest(http.MethodGet, "/api/skills/nonexistent", nil)
	req.SetPathValue("name", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestSkillsHandler_Get_EmptyName(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	req := httptest.NewRequest(http.MethodGet, "/api/skills/", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSkillsHandler_Update(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	body := bytes.NewReader([]byte(`{"enabled": true}`))

	req := httptest.NewRequest(http.MethodPut, "/api/skills/test-skill", body)
	req.SetPathValue("name", "test-skill")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Update() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var skill skills.Skill
	if err := json.Unmarshal(rec.Body.Bytes(), &skill); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if skill.Enabled != true {
		t.Errorf("Enabled = %v, want true", skill.Enabled)
	}
}

func TestSkillsHandler_Update_NotFound(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	body := bytes.NewReader([]byte(`{"enabled": true}`))

	req := httptest.NewRequest(http.MethodPut, "/api/skills/nonexistent", body)
	req.SetPathValue("name", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Update() status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestSkillsHandler_Update_InvalidBody(t *testing.T) {
	loader, _ := setupSkillsTestEnv(t)
	handler := NewSkillsHandler(loader)

	body := bytes.NewReader([]byte(`{"invalid json`))

	req := httptest.NewRequest(http.MethodPut, "/api/skills/test-skill", body)
	req.SetPathValue("name", "test-skill")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Update() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
