package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/user/deer-flow-go/internal/model"
	"github.com/user/deer-flow-go/internal/store"
)

type RunsJoinHandler struct {
	runStore *store.RunStore
}

func NewRunsJoinHandler(runStore *store.RunStore) *RunsJoinHandler {
	return &RunsJoinHandler{runStore: runStore}
}

type JoinResponse struct {
	Run *model.Run `json:"run"`
}

func (h *RunsJoinHandler) Join(w http.ResponseWriter, r *http.Request) {
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

	timeout := parseTimeout(r.URL.Query().Get("timeout"))
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	run, err := h.waitForCompletion(ctx, threadID, runID)
	if err != nil {
		if err == store.ErrRunNotFound {
			writeError(w, http.StatusNotFound, "run not found", "")
			return
		}
		if ctx.Err() == context.DeadlineExceeded {
			writeError(w, http.StatusRequestTimeout, "wait timeout", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to wait for run", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, JoinResponse{Run: run})
}

func (h *RunsJoinHandler) waitForCompletion(ctx context.Context, threadID, runID string) (*model.Run, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		run, err := h.runStore.Get(threadID, runID)
		if err != nil {
			return nil, err
		}

		if isTerminalStatus(run.Status) {
			return run, nil
		}

		select {
		case <-ctx.Done():
			return run, ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

func isTerminalStatus(status string) bool {
	switch status {
	case model.RunStatusSuccess, model.RunStatusError, model.RunStatusCancelled:
		return true
	default:
		return false
	}
}

func parseTimeout(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
