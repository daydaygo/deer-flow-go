package model

import "time"

type ThreadState struct {
	ThreadID  string         `json:"thread_id"`
	Messages  []Message      `json:"messages,omitempty"`
	Title     string         `json:"title,omitempty"`
	Artifacts []Artifact     `json:"artifacts,omitempty"`
	Context   map[string]any `json:"context,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Thread struct {
	ThreadID  string         `json:"thread_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}
