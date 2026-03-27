package mcp

import (
	"context"
	"fmt"
	"sync"
)

type DiscoveryResult struct {
	ServerName string
	Tools      []Tool
	Error      error
}

type ToolDiscovery struct {
	manager *Manager
	mu      sync.RWMutex
	cache   map[string][]Tool
}

func NewToolDiscovery(manager *Manager) *ToolDiscovery {
	return &ToolDiscovery{
		manager: manager,
		cache:   make(map[string][]Tool),
	}
}

func (d *ToolDiscovery) DiscoverAll(ctx context.Context) []DiscoveryResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	results := make([]DiscoveryResult, 0)
	clients := d.manager.ListClients()

	for _, name := range clients {
		client := d.manager.GetClient(name)
		if client == nil {
			results = append(results, DiscoveryResult{
				ServerName: name,
				Error:      fmt.Errorf("client not found"),
			})
			continue
		}

		if !client.IsConnected() {
			if err := client.Connect(ctx); err != nil {
				results = append(results, DiscoveryResult{
					ServerName: name,
					Error:      fmt.Errorf("failed to connect: %w", err),
				})
				continue
			}
		}

		tools, err := client.ListTools(ctx)
		if err != nil {
			results = append(results, DiscoveryResult{
				ServerName: name,
				Error:      err,
			})
			continue
		}

		d.cache[name] = tools
		results = append(results, DiscoveryResult{
			ServerName: name,
			Tools:      tools,
		})
	}

	return results
}

func (d *ToolDiscovery) DiscoverServer(ctx context.Context, name string) (*DiscoveryResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	client := d.manager.GetClient(name)
	if client == nil {
		return nil, fmt.Errorf("client %s not found", name)
	}

	if !client.IsConnected() {
		if err := client.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %w", name, err)
		}
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from %s: %w", name, err)
	}

	d.cache[name] = tools

	return &DiscoveryResult{
		ServerName: name,
		Tools:      tools,
	}, nil
}

func (d *ToolDiscovery) GetCachedTools(serverName string) []Tool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.cache[serverName]
}

func (d *ToolDiscovery) GetAllCachedTools() map[string][]Tool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string][]Tool, len(d.cache))
	for k, v := range d.cache {
		result[k] = v
	}
	return result
}

func (d *ToolDiscovery) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string][]Tool)
}

func ConvertToToolMap(results []DiscoveryResult) map[string]Tool {
	toolMap := make(map[string]Tool)

	for _, result := range results {
		if result.Error != nil {
			continue
		}
		for _, tool := range result.Tools {
			key := fmt.Sprintf("%s_%s", result.ServerName, tool.Name)
			toolMap[key] = tool
		}
	}

	return toolMap
}
