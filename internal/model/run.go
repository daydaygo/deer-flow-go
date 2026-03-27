package model

import "time"

type Run struct {
	RunID       string         `json:"run_id"`
	ThreadID    string         `json:"thread_id"`
	AssistantID string         `json:"assistant_id"`
	Status      string         `json:"status"`
	Input       map[string]any `json:"input,omitempty"`
	Output      map[string]any `json:"output,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type RunCreateRequest struct {
	AssistantID string         `json:"assistant_id"`
	Input       map[string]any `json:"input,omitempty"`
	Config      map[string]any `json:"config,omitempty"`
}

const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusSuccess   = "success"
	RunStatusError     = "error"
	RunStatusCancelled = "cancelled"
)
