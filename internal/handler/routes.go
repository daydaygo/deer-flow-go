package handler

import (
	"net/http"

	"github.com/user/deer-flow-go/internal/config"
)

type Router struct {
	health *HealthHandler
	models *ModelsHandler
}

func NewRouter(cfg *config.Config) *Router {
	return &Router{
		health: NewHealthHandler(),
		models: NewModelsHandler(cfg),
	}
}

func (r *Router) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", r.health.Health)
	mux.HandleFunc("GET /api/models", r.models.List)
	mux.HandleFunc("GET /api/models/{name}", r.models.Get)
}
