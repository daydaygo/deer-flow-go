package mcp

import (
	"context"
	"fmt"
	"sync"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/deer-flow-go/internal/config"
)

type Client struct {
	name      string
	cfg       config.MCPServerConfig
	mcpClient *mcpclient.Client
	mu        sync.RWMutex
	connected bool
}

func NewClient(name string, cfg config.MCPServerConfig) *Client {
	return &Client{
		name: name,
		cfg:  cfg,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	var mcpClient *mcpclient.Client
	var err error

	switch c.cfg.Type {
	case "stdio":
		mcpClient, err = c.connectStdio()
	case "sse":
		mcpClient, err = c.connectSSE()
	case "http":
		mcpClient, err = c.connectHTTP()
	default:
		return fmt.Errorf("unsupported MCP server type: %s", c.cfg.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to MCP server %s: %w", c.name, err)
	}

	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "deer-flow-go",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}

	if _, err := mcpClient.Initialize(ctx, initRequest); err != nil {
		_ = mcpClient.Close()
		return fmt.Errorf("failed to initialize MCP server %s: %w", c.name, err)
	}

	c.mcpClient = mcpClient
	c.connected = true
	return nil
}

func (c *Client) connectStdio() (*mcpclient.Client, error) {
	if c.cfg.Command == "" {
		return nil, fmt.Errorf("command is required for stdio transport")
	}

	env := make([]string, 0, len(c.cfg.Env))
	for k, v := range c.cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return mcpclient.NewStdioMCPClient(c.cfg.Command, env, c.cfg.Args...)
}

func (c *Client) connectSSE() (*mcpclient.Client, error) {
	if c.cfg.URL == "" {
		return nil, fmt.Errorf("url is required for sse transport")
	}

	opts := make([]transport.ClientOption, 0)
	if len(c.cfg.Headers) > 0 {
		opts = append(opts, transport.WithHeaders(c.cfg.Headers))
	}

	return mcpclient.NewSSEMCPClient(c.cfg.URL, opts...)
}

func (c *Client) connectHTTP() (*mcpclient.Client, error) {
	if c.cfg.URL == "" {
		return nil, fmt.Errorf("url is required for http transport")
	}

	opts := make([]transport.StreamableHTTPCOption, 0)
	if len(c.cfg.Headers) > 0 {
		opts = append(opts, transport.WithHTTPHeaders(c.cfg.Headers))
	}

	return mcpclient.NewStreamableHttpClient(c.cfg.URL, opts...)
}

func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.mcpClient == nil {
		return nil, fmt.Errorf("client not connected")
	}

	result, err := c.mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make([]Tool, 0, len(result.Tools))
	for _, t := range result.Tools {
		tools = append(tools, Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	return tools, nil
}

func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolCallResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.mcpClient == nil {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}

	result, err := c.mcpClient.CallTool(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", name, err)
	}

	toolResult := &ToolCallResult{
		IsError: result.IsError,
		Content: make([]ToolContent, 0, len(result.Content)),
	}

	for _, content := range result.Content {
		tc := ToolContent{}
		if textContent, ok := mcp.AsTextContent(content); ok {
			tc.Type = "text"
			tc.Text = textContent.Text
		}
		toolResult.Content = append(toolResult.Content, tc)
	}

	return toolResult, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.mcpClient != nil {
		if err := c.mcpClient.Close(); err != nil {
			return err
		}
	}
	c.connected = false
	return nil
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *Client) Name() string {
	return c.name
}

type Manager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

func (m *Manager) GetClient(name string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[name]
}

func (m *Manager) AddClient(name string, cfg config.MCPServerConfig) *Client {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.clients[name]; ok {
		return existing
	}

	c := NewClient(name, cfg)
	m.clients[name] = c
	return c
}

func (m *Manager) RemoveClient(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.clients[name]; ok {
		if err := c.Close(); err != nil {
			return err
		}
		delete(m.clients, name)
	}
	return nil
}

func (m *Manager) ListClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}

func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, c := range m.clients {
		if err := c.Close(); err != nil {
			lastErr = fmt.Errorf("error closing client %s: %w", name, err)
		}
	}
	m.clients = make(map[string]*Client)
	return lastErr
}
