package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/deer-flow-go/internal/agent"
	"github.com/user/deer-flow-go/internal/model"
	"github.com/user/deer-flow-go/internal/store"
)

type RunsStreamHandler struct {
	runStore *store.RunStore
	engine   *agent.Engine
}

func NewRunsStreamHandler(runStore *store.RunStore, engine *agent.Engine) *RunsStreamHandler {
	return &RunsStreamHandler{
		runStore: runStore,
		engine:   engine,
	}
}

func (h *RunsStreamHandler) Stream(w http.ResponseWriter, r *http.Request) {
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

	sse := agent.NewSSEWriter(w)
	if sse == nil {
		writeError(w, http.StatusInternalServerError, "streaming not supported", "")
		return
	}

	_, _ = h.runStore.Update(threadID, run.RunID, func(r *model.Run) {
		r.Status = model.RunStatusRunning
	})

	inputMessage := extractInputMessage(req.Input)

	initialValues := map[string]interface{}{
		"messages": []model.Message{},
	}
	if err := sse.WriteEvent("values", initialValues); err != nil {
		return
	}

	response, err := h.engine.Invoke(r.Context(), threadID, inputMessage)
	if err != nil {
		_, _ = h.runStore.Update(threadID, run.RunID, func(r *model.Run) {
			r.Status = model.RunStatusError
			r.Output = map[string]any{"error": err.Error()}
		})
		return
	}

	messagesTuple := map[string]interface{}{
		"message": response,
	}
	if err := sse.WriteEvent("messages-tuple", messagesTuple); err != nil {
		return
	}

	_, _ = h.runStore.Update(threadID, run.RunID, func(r *model.Run) {
		r.Status = model.RunStatusSuccess
		r.Output = map[string]any{
			"messages": []model.Message{*response},
		}
	})

	sse.Close()
}

func extractInputMessage(input map[string]any) string {
	if input == nil {
		return ""
	}

	if msg, ok := input["message"].(string); ok {
		return msg
	}

	if messages, ok := input["messages"].([]interface{}); ok && len(messages) > 0 {
		if lastMsg, ok := messages[len(messages)-1].(map[string]interface{}); ok {
			if content, ok := lastMsg["content"].(string); ok {
				return content
			}
		}
	}

	return ""
}
