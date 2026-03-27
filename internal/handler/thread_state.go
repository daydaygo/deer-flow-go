package handler

import (
	"net/http"
	"time"

	"github.com/user/deer-flow-go/internal/model"
	"github.com/user/deer-flow-go/internal/store"
)

type ThreadStateHandler struct {
	threadStore *store.ThreadStore
	runStore    *store.RunStore
}

func NewThreadStateHandler(threadStore *store.ThreadStore, runStore *store.RunStore) *ThreadStateHandler {
	return &ThreadStateHandler{
		threadStore: threadStore,
		runStore:    runStore,
	}
}

type ThreadStateResponse struct {
	ThreadID  string         `json:"thread_id"`
	Values    map[string]any `json:"values"`
	Next      []string       `json:"next,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (h *ThreadStateHandler) Get(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id is required", "")
		return
	}

	thread, err := h.threadStore.Get(threadID)
	if err != nil {
		if err == store.ErrThreadNotFound {
			writeError(w, http.StatusNotFound, "thread not found", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get thread", err.Error())
		return
	}

	runs, err := h.runStore.List(threadID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list runs", err.Error())
		return
	}

	values := map[string]any{
		"messages":  []model.Message{},
		"artifacts": []model.Artifact{},
	}

	var latestRun *model.Run
	for _, run := range runs {
		if latestRun == nil || run.UpdatedAt.After(latestRun.UpdatedAt) {
			latestRun = run
		}
	}

	var next []string
	if latestRun != nil {
		if latestRun.Status == model.RunStatusPending || latestRun.Status == model.RunStatusRunning {
			next = []string{"agent"}
		}
		if latestRun.Output != nil {
			for k, v := range latestRun.Output {
				values[k] = v
			}
		}
	}

	response := ThreadStateResponse{
		ThreadID:  threadID,
		Values:    values,
		Next:      next,
		CreatedAt: thread.CreatedAt,
		UpdatedAt: thread.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, response)
}
