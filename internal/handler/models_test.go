package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/deer-flow-go/internal/config"
)

func TestModelsHandler_List(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{
				Name:             "gpt-4o",
				DisplayName:      "GPT-4o",
				SupportsThinking: false,
				SupportsVision:   true,
			},
			{
				Name:             "claude-3.5-sonnet",
				DisplayName:      "Claude 3.5 Sonnet",
				SupportsThinking: true,
				SupportsVision:   true,
			},
		},
	}

	handler := NewModelsHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var response ModelsListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(response.Models) != 2 {
		t.Errorf("len(Models) = %d, want 2", len(response.Models))
	}

	if response.Models[0].Name != "gpt-4o" {
		t.Errorf("Models[0].Name = %q, want %q", response.Models[0].Name, "gpt-4o")
	}
	if response.Models[0].DisplayName != "GPT-4o" {
		t.Errorf("Models[0].DisplayName = %q, want %q", response.Models[0].DisplayName, "GPT-4o")
	}
	if response.Models[0].SupportsThinking != false {
		t.Errorf("Models[0].SupportsThinking = %v, want false", response.Models[0].SupportsThinking)
	}
	if response.Models[0].SupportsVision != true {
		t.Errorf("Models[0].SupportsVision = %v, want true", response.Models[0].SupportsVision)
	}
}

func TestModelsHandler_List_Empty(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{},
	}

	handler := NewModelsHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response ModelsListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(response.Models) != 0 {
		t.Errorf("len(Models) = %d, want 0", len(response.Models))
	}
}

func TestModelsHandler_Get_Found(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{
				Name:             "gpt-4o",
				DisplayName:      "GPT-4o",
				SupportsThinking: false,
				SupportsVision:   true,
			},
			{
				Name:             "claude-3.5-sonnet",
				DisplayName:      "Claude 3.5 Sonnet",
				SupportsThinking: true,
				SupportsVision:   true,
			},
		},
	}

	handler := NewModelsHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/models/claude-3.5-sonnet", nil)
	req.SetPathValue("name", "claude-3.5-sonnet")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var response ModelResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Name != "claude-3.5-sonnet" {
		t.Errorf("Name = %q, want %q", response.Name, "claude-3.5-sonnet")
	}
	if response.DisplayName != "Claude 3.5 Sonnet" {
		t.Errorf("DisplayName = %q, want %q", response.DisplayName, "Claude 3.5 Sonnet")
	}
	if response.SupportsThinking != true {
		t.Errorf("SupportsThinking = %v, want true", response.SupportsThinking)
	}
	if response.SupportsVision != true {
		t.Errorf("SupportsVision = %v, want true", response.SupportsVision)
	}
}

func TestModelsHandler_Get_NotFound(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{
				Name:        "gpt-4o",
				DisplayName: "GPT-4o",
			},
		},
	}

	handler := NewModelsHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent", nil)
	req.SetPathValue("name", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response["error"] != "model not found" {
		t.Errorf("error = %q, want %q", response["error"], "model not found")
	}
}

func TestModelsHandler_Get_EmptyName(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{},
	}

	handler := NewModelsHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/models/", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
