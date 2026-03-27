package handler

import "net/http"

type Router struct {
	health *HealthHandler
}

func NewRouter() *Router {
	return &Router{
		health: NewHealthHandler(),
	}
}

func (r *Router) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", r.health.Health)
}
