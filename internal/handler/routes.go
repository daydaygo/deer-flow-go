package handler

import (
	"net/http"

	"github.com/user/deer-flow-go/internal/config"
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
}

func NewRouter(cfg *config.Config) *Router {
	memoryStore := store.NewMemoryStore(cfg.Memory.StoragePath)
	threadStore := store.NewThreadStore(cfg.Storage.DataDir)
	runStore := store.NewRunStore(cfg.Storage.DataDir)

	return &Router{
		health:      NewHealthHandler(),
		models:      NewModelsHandler(cfg),
		memory:      NewMemoryHandler(memoryStore),
		threads:     NewThreadsHandler(threadStore),
		runs:        NewRunsHandler(runStore),
		threadState: NewThreadStateHandler(threadStore, runStore),
		runsJoin:    NewRunsJoinHandler(runStore),
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
}
