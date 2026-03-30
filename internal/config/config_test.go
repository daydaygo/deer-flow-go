package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  name: test-server
  host: 0.0.0.0
  port: 8001

models:
  - name: gpt-4o
    display_name: GPT-4o
    use: openai
    api_key: test-key
    supports_thinking: false
    supports_vision: true

memory:
  enabled: true
  storage_path: .deer-flow/memory.json
  injection_enabled: true
  max_injection_tokens: 2000

storage:
  data_dir: .deer-flow
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Name != "test-server" {
		t.Errorf("Server.Name = %q, want %q", cfg.Server.Name, "test-server")
	}

	if len(cfg.Models) != 1 {
		t.Errorf("len(Models) = %d, want 1", len(cfg.Models))
	}

	if cfg.Models[0].Name != "gpt-4o" {
		t.Errorf("Models[0].Name = %q, want %q", cfg.Models[0].Name, "gpt-4o")
	}
}

func TestLoadWithEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  name: test-server
  host: 0.0.0.0
  port: 8001

models:
  - name: gpt-4o
    display_name: GPT-4o
    use: openai
    api_key: $TEST_API_KEY
    supports_thinking: false
    supports_vision: true

memory:
  enabled: true
  storage_path: .deer-flow/memory.json
  injection_enabled: true
  max_injection_tokens: 2000

storage:
  data_dir: .deer-flow
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	os.Setenv("TEST_API_KEY", "env-api-key")
	defer os.Unsetenv("TEST_API_KEY")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Models[0].APIKey != "env-api-key" {
		t.Errorf("Models[0].APIKey = %q, want %q", cfg.Models[0].APIKey, "env-api-key")
	}
}

func TestMustLoad(t *testing.T) {
	t.Run("panics on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustLoad() did not panic")
			}
		}()
		MustLoad("/nonexistent/path/config.yaml")
	})

	t.Run("returns config on success", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		configContent := `
server:
  name: test-server
  host: 0.0.0.0
  port: 8001

models: []

memory:
  enabled: false
  storage_path: ""
  injection_enabled: false
  max_injection_tokens: 0

storage:
  data_dir: ""
`
		if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		cfg := MustLoad(configPath)
		if cfg == nil {
			t.Error("MustLoad() returned nil")
		}
	})
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  name: test-server
  host: 0.0.0.0
  port: 8001

models: []

memory:
  enabled: false
  storage_path: ""
  injection_enabled: false
  max_injection_tokens: 0

storage:
  data_dir: ""
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if Get() != cfg {
		t.Error("Get() did not return the loaded config")
	}
}
