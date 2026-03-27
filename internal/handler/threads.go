package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/deer-flow-go/internal/model"
	"github.com/user/deer-flow-go/internal/store"
)

type ThreadsHandler struct {
	threadStore *store.ThreadStore
}

func NewThreadsHandler(threadStore *store.ThreadStore) *ThreadsHandler {
	return &ThreadsHandler{threadStore: threadStore}
}

func (h *ThreadsHandler) Create(w http.ResponseWriter, r *http.Request) {
	thread, err := h.threadStore.Create()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create thread", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, thread)
}

func (h *ThreadsHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, thread)
}

func (h *ThreadsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id is required", "")
		return
	}

	if err := h.threadStore.Delete(threadID); err != nil {
		if err == store.ErrThreadNotFound {
			writeError(w, http.StatusNotFound, "thread not found", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete thread", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:  msg,
		Detail: detail,
	})
}
