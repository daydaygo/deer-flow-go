package mcp

import (
	"encoding/json"
	"testing"

	"github.com/user/deer-flow-go/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := config.MCPServerConfig{
		Enabled: true,
		Type:    "stdio",
		Command: "test-command",
		Args:    []string{"--arg1", "--arg2"},
		Env:     map[string]string{"ENV1": "value1"},
	}

	client := NewClient("test-server", cfg)

	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.Name() != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", client.Name())
	}

	if client.IsConnected() {
		t.Error("expected client to not be connected initially")
	}
}

func TestClientTypes(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.MCPServerConfig
		expectError bool
	}{
		{
			name: "stdio missing command",
			cfg: config.MCPServerConfig{
				Enabled: true,
				Type:    "stdio",
			},
			expectError: true,
		},
		{
			name: "sse missing url",
			cfg: config.MCPServerConfig{
				Enabled: true,
				Type:    "sse",
			},
			expectError: true,
		},
		{
			name: "http missing url",
			cfg: config.MCPServerConfig{
				Enabled: true,
				Type:    "http",
			},
			expectError: true,
		},
		{
			name: "unknown type",
			cfg: config.MCPServerConfig{
				Enabled: true,
				Type:    "unknown",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("test", tt.cfg)
			err := client.Connect(nil)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("expected manager to be created")
	}

	cfg := config.MCPServerConfig{
		Enabled: true,
		Type:    "stdio",
		Command: "test",
	}

	client1 := manager.AddClient("server1", cfg)
	if client1 == nil {
		t.Error("expected client to be added")
	}

	client1Again := manager.AddClient("server1", cfg)
	if client1Again != client1 {
		t.Error("expected same client for duplicate add")
	}

	client2 := manager.AddClient("server2", cfg)
	if client2 == nil {
		t.Error("expected client2 to be added")
	}

	clients := manager.ListClients()
	if len(clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(clients))
	}

	retrieved := manager.GetClient("server1")
	if retrieved != client1 {
		t.Error("expected to retrieve same client")
	}

	err := manager.RemoveClient("server1")
	if err != nil {
		t.Errorf("unexpected error removing client: %v", err)
	}

	clients = manager.ListClients()
	if len(clients) != 1 {
		t.Errorf("expected 1 client after removal, got %d", len(clients))
	}

	retrieved = manager.GetClient("server1")
	if retrieved != nil {
		t.Error("expected nil for removed client")
	}

	err = manager.CloseAll()
	if err != nil {
		t.Errorf("unexpected error closing all: %v", err)
	}

	clients = manager.ListClients()
	if len(clients) != 0 {
		t.Errorf("expected 0 clients after close all, got %d", len(clients))
	}
}

func TestToolDiscovery(t *testing.T) {
	manager := NewManager()
	discovery := NewToolDiscovery(manager)

	if discovery == nil {
		t.Fatal("expected discovery to be created")
	}

	results := discovery.DiscoverAll(nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty manager, got %d", len(results))
	}

	cached := discovery.GetCachedTools("nonexistent")
	if cached != nil {
		t.Error("expected nil for nonexistent cached tools")
	}

	allCached := discovery.GetAllCachedTools()
	if len(allCached) != 0 {
		t.Errorf("expected empty cache, got %d items", len(allCached))
	}

	discovery.ClearCache()
}

func TestConvertToToolMap(t *testing.T) {
	results := []DiscoveryResult{
		{
			ServerName: "server1",
			Tools: []Tool{
				{Name: "tool1", Description: "desc1"},
				{Name: "tool2", Description: "desc2"},
			},
		},
		{
			ServerName: "server2",
			Error:      nil,
			Tools: []Tool{
				{Name: "tool3", Description: "desc3"},
			},
		},
		{
			ServerName: "server3",
			Error:      nil,
			Tools:      nil,
		},
	}

	toolMap := ConvertToToolMap(results)

	if len(toolMap) != 3 {
		t.Errorf("expected 3 tools in map, got %d", len(toolMap))
	}

	if _, ok := toolMap["server1_tool1"]; !ok {
		t.Error("expected server1_tool1 in map")
	}

	if _, ok := toolMap["server2_tool3"]; !ok {
		t.Error("expected server2_tool3 in map")
	}
}

func TestMCPConfigJSON(t *testing.T) {
	cfg := MCPConfig{
		Servers: map[string]MCPServerConfig{
			"server1": {
				Enabled: true,
				Type:    "stdio",
				Command: "test-cmd",
				Args:    []string{"arg1"},
				Env:     map[string]string{"KEY": "VAL"},
			},
			"server2": {
				Enabled: true,
				Type:    "sse",
				URL:     "http://localhost:8080",
				Headers: map[string]string{"Authorization": "Bearer token"},
			},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	var unmarshaled MCPConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if len(unmarshaled.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(unmarshaled.Servers))
	}

	s1 := unmarshaled.Servers["server1"]
	if s1.Type != "stdio" {
		t.Errorf("expected type 'stdio', got '%s'", s1.Type)
	}
	if s1.Command != "test-cmd" {
		t.Errorf("expected command 'test-cmd', got '%s'", s1.Command)
	}

	s2 := unmarshaled.Servers["server2"]
	if s2.Type != "sse" {
		t.Errorf("expected type 'sse', got '%s'", s2.Type)
	}
	if s2.URL != "http://localhost:8080" {
		t.Errorf("expected url 'http://localhost:8080', got '%s'", s2.URL)
	}
}

func TestToolJSON(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type":        "string",
					"description": "First parameter",
				},
			},
			"required": []string{"param1"},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("failed to marshal tool: %v", err)
	}

	var unmarshaled Tool
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal tool: %v", err)
	}

	if unmarshaled.Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", unmarshaled.Name)
	}
	if unmarshaled.Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got '%s'", unmarshaled.Description)
	}
}

func TestToolCallResultJSON(t *testing.T) {
	result := ToolCallResult{
		IsError: false,
		Content: []ToolContent{
			{Type: "text", Text: "Hello"},
			{Type: "image", Data: "base64data", MimeType: "image/png"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	var unmarshaled ToolCallResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if unmarshaled.IsError != false {
		t.Errorf("expected IsError false, got %v", unmarshaled.IsError)
	}
	if len(unmarshaled.Content) != 2 {
		t.Errorf("expected 2 content items, got %d", len(unmarshaled.Content))
	}
	if unmarshaled.Content[0].Text != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", unmarshaled.Content[0].Text)
	}
}
