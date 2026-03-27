package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageJSONMarshaling(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	msg := Message{
		ID:        "msg-123",
		Role:      "user",
		Content:   "Hello, world!",
		CreatedAt: now,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ID != msg.ID {
		t.Errorf("ID = %q, want %q", unmarshaled.ID, msg.ID)
	}
	if unmarshaled.Role != msg.Role {
		t.Errorf("Role = %q, want %q", unmarshaled.Role, msg.Role)
	}
	if unmarshaled.Content != msg.Content {
		t.Errorf("Content = %q, want %q", unmarshaled.Content, msg.Content)
	}
	if !unmarshaled.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", unmarshaled.CreatedAt, now)
	}
}

func TestMessageWithToolCalls(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	msg := Message{
		ID:      "msg-456",
		Role:    "assistant",
		Content: "Let me help you with that.",
		ToolCalls: []ToolCall{
			{ID: "call-1", Name: "search", Arguments: `{"query": "test"}`},
		},
		CreatedAt: now,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(unmarshaled.ToolCalls) != 1 {
		t.Fatalf("len(ToolCalls) = %d, want 1", len(unmarshaled.ToolCalls))
	}
	if unmarshaled.ToolCalls[0].ID != "call-1" {
		t.Errorf("ToolCalls[0].ID = %q, want %q", unmarshaled.ToolCalls[0].ID, "call-1")
	}
	if unmarshaled.ToolCalls[0].Name != "search" {
		t.Errorf("ToolCalls[0].Name = %q, want %q", unmarshaled.ToolCalls[0].Name, "search")
	}
}

func TestToolResultJSONMarshaling(t *testing.T) {
	result := ToolResult{
		ToolID:  "call-1",
		Content: "Search result...",
		IsError: false,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ToolResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ToolID != result.ToolID {
		t.Errorf("ToolID = %q, want %q", unmarshaled.ToolID, result.ToolID)
	}
	if unmarshaled.Content != result.Content {
		t.Errorf("Content = %q, want %q", unmarshaled.Content, result.Content)
	}
}

func TestThreadStateJSONMarshaling(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	state := ThreadState{
		ThreadID:  "thread-123",
		Title:     "Test Thread",
		CreatedAt: now,
		UpdatedAt: now,
		Messages: []Message{
			{ID: "msg-1", Role: "user", Content: "Hi", CreatedAt: now},
		},
		Artifacts: []Artifact{
			{ID: "art-1", Name: "test.txt", Path: "/tmp/test.txt", MimeType: "text/plain", Size: 100},
		},
		Context: map[string]any{"key": "value"},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ThreadState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ThreadID != state.ThreadID {
		t.Errorf("ThreadID = %q, want %q", unmarshaled.ThreadID, state.ThreadID)
	}
	if unmarshaled.Title != state.Title {
		t.Errorf("Title = %q, want %q", unmarshaled.Title, state.Title)
	}
	if len(unmarshaled.Messages) != 1 {
		t.Errorf("len(Messages) = %d, want 1", len(unmarshaled.Messages))
	}
	if len(unmarshaled.Artifacts) != 1 {
		t.Errorf("len(Artifacts) = %d, want 1", len(unmarshaled.Artifacts))
	}
}

func TestThreadJSONMarshaling(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	thread := Thread{
		ThreadID:  "thread-456",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  map[string]any{"source": "api"},
	}

	data, err := json.Marshal(thread)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Thread
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ThreadID != thread.ThreadID {
		t.Errorf("ThreadID = %q, want %q", unmarshaled.ThreadID, thread.ThreadID)
	}
}

func TestErrorResponseJSONMarshaling(t *testing.T) {
	errResp := ErrorResponse{
		Error:  "Not found",
		Code:   "NOT_FOUND",
		Detail: "Resource does not exist",
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Error != errResp.Error {
		t.Errorf("Error = %q, want %q", unmarshaled.Error, errResp.Error)
	}
	if unmarshaled.Code != errResp.Code {
		t.Errorf("Code = %q, want %q", unmarshaled.Code, errResp.Code)
	}
}

func TestArtifactJSONMarshaling(t *testing.T) {
	artifact := Artifact{
		ID:       "art-789",
		Name:     "document.pdf",
		Path:     "/data/artifacts/document.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Artifact
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ID != artifact.ID {
		t.Errorf("ID = %q, want %q", unmarshaled.ID, artifact.ID)
	}
	if unmarshaled.Size != artifact.Size {
		t.Errorf("Size = %d, want %d", unmarshaled.Size, artifact.Size)
	}
}

func TestMessageOmitEmptyFields(t *testing.T) {
	msg := Message{
		ID:        "msg-empty",
		Role:      "user",
		Content:   "test",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := raw["tool_calls"]; exists {
		t.Error("tool_calls should be omitted when empty")
	}
	if _, exists := raw["tool_id"]; exists {
		t.Error("tool_id should be omitted when empty")
	}
	if _, exists := raw["name"]; exists {
		t.Error("name should be omitted when empty")
	}
}

func TestThreadStateOmitEmptyFields(t *testing.T) {
	now := time.Now()
	state := ThreadState{
		ThreadID:  "thread-empty",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := raw["messages"]; exists {
		t.Error("messages should be omitted when empty")
	}
	if _, exists := raw["title"]; exists {
		t.Error("title should be omitted when empty")
	}
	if _, exists := raw["artifacts"]; exists {
		t.Error("artifacts should be omitted when empty")
	}
	if _, exists := raw["context"]; exists {
		t.Error("context should be omitted when empty")
	}
}
