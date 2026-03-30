package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/deer-flow-go/internal/mcp"
)

func TestMCPConfigStoreLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	store := NewMCPConfigStore(configPath)

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected config to be returned")
	}

	if cfg.Servers == nil {
		t.Error("expected Servers map to be initialized")
	}
}

func TestMCPConfigStoreSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	store := NewMCPConfigStore(configPath)

	cfg := &mcp.MCPConfig{
		Servers: map[string]mcp.MCPServerConfig{
			"test-server": {
				Enabled: true,
				Type:    "stdio",
				Command: "test-cmd",
			},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	var loaded mcp.MCPConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal saved config: %v", err)
	}

	if len(loaded.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(loaded.Servers))
	}

	if loaded.Servers["test-server"].Command != "test-cmd" {
		t.Errorf("expected command 'test-cmd', got '%s'", loaded.Servers["test-server"].Command)
	}
}

func TestMCPConfigStoreRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	store := NewMCPConfigStore(configPath)

	cfg := &mcp.MCPConfig{
		Servers: map[string]mcp.MCPServerConfig{
			"server1": {
				Enabled: true,
				Type:    "stdio",
				Command: "cmd1",
				Args:    []string{"arg1"},
				Env:     map[string]string{"KEY": "VAL"},
			},
			"server2": {
				Enabled: true,
				Type:    "sse",
				URL:     "http://localhost:8080/sse",
				Headers: map[string]string{"Auth": "token"},
			},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if len(loaded.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(loaded.Servers))
	}

	s1 := loaded.Servers["server1"]
	if s1.Type != "stdio" || s1.Command != "cmd1" {
		t.Errorf("server1 data mismatch: type=%s, cmd=%s", s1.Type, s1.Command)
	}

	s2 := loaded.Servers["server2"]
	if s2.Type != "sse" || s2.URL != "http://localhost:8080/sse" {
		t.Errorf("server2 data mismatch: type=%s, url=%s", s2.Type, s2.URL)
	}
}

func TestMCPHandlerGetConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/config", nil)
	rec := httptest.NewRecorder()

	handler.GetConfig(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var cfg mcp.MCPConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if cfg.Servers == nil {
		t.Error("expected Servers to be initialized")
	}
}

func TestMCPHandlerUpdateConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	cfg := mcp.MCPConfig{
		Servers: map[string]mcp.MCPServerConfig{
			"test-server": {
				Enabled: true,
				Type:    "stdio",
				Command: "test-command",
			},
		},
	}

	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/mcp/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.UpdateConfig(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response mcp.MCPConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Servers) != 1 {
		t.Errorf("expected 1 server in response, got %d", len(response.Servers))
	}
}

func TestMCPHandlerUpdateConfigInvalidBody(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	req := httptest.NewRequest(http.MethodPut, "/api/mcp/config", bytes.NewReader([]byte("invalid json")))
	rec := httptest.NewRecorder()

	handler.UpdateConfig(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMCPHandlerListTools(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/tools", nil)
	rec := httptest.NewRecorder()

	handler.ListTools(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var results []mcp.DiscoveryResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty manager, got %d", len(results))
	}
}

func TestMCPHandlerCallToolMissingServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp/servers//tools/", nil)
	rec := httptest.NewRecorder()

	handler.CallTool(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing server, got %d", rec.Code)
	}
}

func TestMCPHandlerCallToolMissingTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/test-server/tools/", nil)
	req.SetPathValue("server", "test-server")
	rec := httptest.NewRecorder()

	handler.CallTool(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing tool, got %d", rec.Code)
	}
}

func TestMCPHandlerCallToolNonexistentServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	handler := NewMCPHandler(configPath)

	args := map[string]any{"param": "value"}
	body, _ := json.Marshal(args)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/nonexistent/tools/test-tool", bytes.NewReader(body))
	req.SetPathValue("server", "nonexistent")
	req.SetPathValue("tool", "test-tool")
	rec := httptest.NewRecorder()

	handler.CallTool(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for nonexistent server, got %d", rec.Code)
	}
}
