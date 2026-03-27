package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/deer-flow-go/internal/model"
	"github.com/user/deer-flow-go/internal/store"
)

type RunsHandler struct {
	runStore *store.RunStore
}

func NewRunsHandler(runStore *store.RunStore) *RunsHandler {
	return &RunsHandler{runStore: runStore}
}

func (h *RunsHandler) Create(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id is required", "")
		return
	}

	var req model.RunCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.AssistantID == "" {
		req.AssistantID = "agent"
	}

	run, err := h.runStore.Create(threadID, req.AssistantID, req.Input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create run", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (h *RunsHandler) Get(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id is required", "")
		return
	}

	runID := r.PathValue("run_id")
	if runID == "" {
		writeError(w, http.StatusBadRequest, "run_id is required", "")
		return
	}

	run, err := h.runStore.Get(threadID, runID)
	if err != nil {
		if err == store.ErrRunNotFound {
			writeError(w, http.StatusNotFound, "run not found", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get run", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, run)
}
