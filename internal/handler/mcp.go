package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/mcp"
)

type MCPConfigStore struct {
	configPath string
	mu         sync.RWMutex
	cache      *mcp.MCPConfig
}

func NewMCPConfigStore(configPath string) *MCPConfigStore {
	return &MCPConfigStore{
		configPath: configPath,
	}
}

func (s *MCPConfigStore) Load() (*mcp.MCPConfig, error) {
	s.mu.RLock()
	if s.cache != nil {
		defer s.mu.RUnlock()
		return s.cache, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil {
		return s.cache, nil
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &mcp.MCPConfig{
				Servers: make(map[string]mcp.MCPServerConfig),
			}
			s.cache = cfg
			return cfg, nil
		}
		return nil, err
	}

	var cfg mcp.MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Servers == nil {
		cfg.Servers = make(map[string]mcp.MCPServerConfig)
	}

	s.cache = &cfg
	return &cfg, nil
}

func (s *MCPConfigStore) Save(cfg *mcp.MCPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := s.configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, s.configPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	s.cache = cfg
	return nil
}

type MCPHandler struct {
	configStore *MCPConfigStore
	manager     *mcp.Manager
	discovery   *mcp.ToolDiscovery
}

func NewMCPHandler(configPath string) *MCPHandler {
	store := NewMCPConfigStore(configPath)
	manager := mcp.NewManager()
	discovery := mcp.NewToolDiscovery(manager)
	return &MCPHandler{
		configStore: store,
		manager:     manager,
		discovery:   discovery,
	}
}

func (h *MCPHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.configStore.Load()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

func (h *MCPHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg mcp.MCPConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if cfg.Servers == nil {
		cfg.Servers = make(map[string]mcp.MCPServerConfig)
	}

	if err := h.configStore.Save(&cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	for _, name := range h.manager.ListClients() {
		if _, exists := cfg.Servers[name]; !exists {
			_ = h.manager.RemoveClient(name)
		}
	}

	for name, serverCfg := range cfg.Servers {
		if serverCfg.Enabled {
			converted := config.MCPServerConfig{
				Enabled: serverCfg.Enabled,
				Type:    serverCfg.Type,
				Command: serverCfg.Command,
				Args:    serverCfg.Args,
				URL:     serverCfg.URL,
				Env:     serverCfg.Env,
				Headers: serverCfg.Headers,
			}
			h.manager.AddClient(name, converted)
		} else {
			_ = h.manager.RemoveClient(name)
		}
	}

	h.discovery.ClearCache()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

func (h *MCPHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	results := h.discovery.DiscoverAll(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func (h *MCPHandler) CallTool(w http.ResponseWriter, r *http.Request) {
	serverName := r.PathValue("server")
	if serverName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "server name is required"})
		return
	}

	toolName := r.PathValue("tool")
	if toolName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "tool name is required"})
		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		args = make(map[string]interface{})
	}

	client := h.manager.GetClient(serverName)
	if client == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "server not found"})
		return
	}

	if !client.IsConnected() {
		if err := client.Connect(r.Context()); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	}

	result, err := client.CallTool(r.Context(), toolName, args)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
