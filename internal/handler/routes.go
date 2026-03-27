package handler

import (
	"net/http"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/store"
)

type Router struct {
	health *HealthHandler
	models *ModelsHandler
	memory *MemoryHandler
}

func NewRouter(cfg *config.Config) *Router {
	memoryStore := store.NewMemoryStore(cfg.Memory.StoragePath)
	return &Router{
		health: NewHealthHandler(),
		models: NewModelsHandler(cfg),
		memory: NewMemoryHandler(memoryStore),
	}
}

func (r *Router) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", r.health.Health)
	mux.HandleFunc("GET /api/models", r.models.List)
	mux.HandleFunc("GET /api/models/{name}", r.models.Get)
	mux.HandleFunc("GET /api/memory", r.memory.Get)
	mux.HandleFunc("POST /api/memory/reload", r.memory.Reload)
}
