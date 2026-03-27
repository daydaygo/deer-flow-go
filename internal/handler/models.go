package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/deer-flow-go/internal/config"
)

type ModelResponse struct {
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	SupportsThinking bool   `json:"supports_thinking"`
	SupportsVision   bool   `json:"supports_vision"`
}

type ModelsListResponse struct {
	Models []ModelResponse `json:"models"`
}

type ModelsHandler struct {
	config *config.Config
}

func NewModelsHandler(cfg *config.Config) *ModelsHandler {
	return &ModelsHandler{config: cfg}
}

func (h *ModelsHandler) List(w http.ResponseWriter, r *http.Request) {
	models := make([]ModelResponse, 0, len(h.config.Models))
	for _, m := range h.config.Models {
		models = append(models, ModelResponse{
			Name:             m.Name,
			DisplayName:      m.DisplayName,
			SupportsThinking: m.SupportsThinking,
			SupportsVision:   m.SupportsVision,
		})
	}

	response := ModelsListResponse{Models: models}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ModelsHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "model name is required", http.StatusBadRequest)
		return
	}

	for _, m := range h.config.Models {
		if m.Name == name {
			response := ModelResponse{
				Name:             m.Name,
				DisplayName:      m.DisplayName,
				SupportsThinking: m.SupportsThinking,
				SupportsVision:   m.SupportsVision,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "model not found"})
}
