package handler

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/user/deer-flow-go/internal/agent"
	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/llm"
	"github.com/user/deer-flow-go/internal/skills"
	"github.com/user/deer-flow-go/internal/store"
)

type Router struct {
	health      *HealthHandler
	models      *ModelsHandler
	memory      *MemoryHandler
	threads     *ThreadsHandler
	runs        *RunsHandler
	threadState *ThreadStateHandler
	runsJoin    *RunsJoinHandler
	runsStream  *RunsStreamHandler
	skills      *SkillsHandler
	mcp         *MCPHandler
}

func NewRouter(cfg *config.Config, configPath string) *Router {
	memoryStore := store.NewMemoryStore(cfg.Memory.StoragePath)
	threadStore := store.NewThreadStore(cfg.Storage.DataDir)
	runStore := store.NewRunStore(cfg.Storage.DataDir)

	llmFactory := llm.NewFactory(cfg)
	engine := agent.NewEngine(cfg, llmFactory)

	skillsLoader := skills.NewLoader("skills", "extensions_config.json")
	if err := skillsLoader.Load(); err != nil {
		panic(fmt.Sprintf("failed to load skills: %v", err))
	}

	mcpConfigPath := filepath.Join(filepath.Dir(configPath), "extensions_config.json")
	mcpHandler := NewMCPHandler(mcpConfigPath)

	return &Router{
		health:      NewHealthHandler(),
		models:      NewModelsHandler(cfg),
		memory:      NewMemoryHandler(memoryStore),
		threads:     NewThreadsHandler(threadStore),
		runs:        NewRunsHandler(runStore),
		threadState: NewThreadStateHandler(threadStore, runStore),
		runsJoin:    NewRunsJoinHandler(runStore),
		runsStream:  NewRunsStreamHandler(runStore, engine),
		skills:      NewSkillsHandler(skillsLoader),
		mcp:         mcpHandler,
	}
}

func (r *Router) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", r.health.Health)
	mux.HandleFunc("GET /api/models", r.models.List)
	mux.HandleFunc("GET /api/models/{name}", r.models.Get)
	mux.HandleFunc("GET /api/memory", r.memory.Get)
	mux.HandleFunc("POST /api/memory/reload", r.memory.Reload)

	mux.HandleFunc("POST /api/langgraph/threads", r.threads.Create)
	mux.HandleFunc("GET /api/langgraph/threads/{id}", r.threads.Get)
	mux.HandleFunc("DELETE /api/langgraph/threads/{id}", r.threads.Delete)

	mux.HandleFunc("POST /api/langgraph/threads/{id}/runs", r.runs.Create)
	mux.HandleFunc("GET /api/langgraph/threads/{id}/runs/{run_id}", r.runs.Get)

	mux.HandleFunc("GET /api/langgraph/threads/{id}/state", r.threadState.Get)

	mux.HandleFunc("POST /api/langgraph/threads/{id}/runs/{run_id}/join", r.runsJoin.Join)

	mux.HandleFunc("POST /api/langgraph/threads/{id}/runs/stream", r.runsStream.Stream)

	mux.HandleFunc("GET /api/skills", r.skills.List)
	mux.HandleFunc("GET /api/skills/{name}", r.skills.Get)
	mux.HandleFunc("PUT /api/skills/{name}", r.skills.Update)

	mux.HandleFunc("GET /api/mcp/config", r.mcp.GetConfig)
	mux.HandleFunc("PUT /api/mcp/config", r.mcp.UpdateConfig)
	mux.HandleFunc("GET /api/mcp/tools", r.mcp.ListTools)
	mux.HandleFunc("POST /api/mcp/servers/{server}/tools/{tool}", r.mcp.CallTool)
}
