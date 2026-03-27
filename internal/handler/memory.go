package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/deer-flow-go/internal/store"
)

type MemoryHandler struct {
	store *store.MemoryStore
}

func NewMemoryHandler(memoryStore *store.MemoryStore) *MemoryHandler {
	return &MemoryHandler{store: memoryStore}
}

func (h *MemoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	mem, err := h.store.Load()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mem)
}

func (h *MemoryHandler) Reload(w http.ResponseWriter, r *http.Request) {
	mem, err := h.store.Reload()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mem)
}
