package subagentstypes

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	task := NewTask("thread-1", "general-purpose", "test task", "test prompt")

	if task.TaskID == "" {
		t.Error("expected task ID to be generated")
	}
	if task.ThreadID != "thread-1" {
		t.Errorf("expected thread ID 'thread-1', got %s", task.ThreadID)
	}
	if task.SubagentType != "general-purpose" {
		t.Errorf("expected subagent type 'general-purpose', got %s", task.SubagentType)
	}
	if task.Status != StatusPending {
		t.Errorf("expected status 'pending', got %s", task.Status)
	}
	if task.MaxTurns != DefaultMaxTurns {
		t.Errorf("expected max turns %d, got %d", DefaultMaxTurns, task.MaxTurns)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected created at to be set")
	}
	if task.AIMessages == nil {
		t.Error("expected AI messages to be initialized")
	}
}

func TestTask_IsTerminal(t *testing.T) {
	tests := []struct {
		status   Status
		expected bool
	}{
		{StatusPending, false},
		{StatusRunning, false},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusTimeout, true},
	}

	for _, tt := range tests {
		task := &Task{Status: tt.status}
		if task.IsTerminal() != tt.expected {
			t.Errorf("status %s: expected IsTerminal=%v, got %v", tt.status, tt.expected, task.IsTerminal())
		}
	}
}

func TestGenerateTaskID(t *testing.T) {
	id1 := GenerateTaskID()
	id2 := GenerateTaskID()

	if id1 == "" {
		t.Error("expected task ID to be non-empty")
	}
	if len(id1) != 8 {
		t.Errorf("expected task ID length 8, got %d", len(id1))
	}
	if id1 == id2 {
		t.Error("expected different task IDs for consecutive calls")
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusPending != "pending" {
		t.Errorf("expected StatusPending 'pending', got %s", StatusPending)
	}
	if StatusRunning != "running" {
		t.Errorf("expected StatusRunning 'running', got %s", StatusRunning)
	}
	if StatusCompleted != "completed" {
		t.Errorf("expected StatusCompleted 'completed', got %s", StatusCompleted)
	}
	if StatusFailed != "failed" {
		t.Errorf("expected StatusFailed 'failed', got %s", StatusFailed)
	}
	if StatusTimeout != "timeout" {
		t.Errorf("expected StatusTimeout 'timeout', got %s", StatusTimeout)
	}
}

func TestEventTypeConstants(t *testing.T) {
	if EventTypeStarted != "task_started" {
		t.Errorf("expected EventTypeStarted 'task_started', got %s", EventTypeStarted)
	}
	if EventTypeCompleted != "task_completed" {
		t.Errorf("expected EventTypeCompleted 'task_completed', got %s", EventTypeCompleted)
	}
}

func TestTaskEvent(t *testing.T) {
	event := TaskEvent{
		Type:      EventTypeStarted,
		TaskID:    "test-123",
		ThreadID:  "thread-1",
		Subagent:  "general-purpose",
		Content:   "test content",
		Timestamp: time.Now(),
	}

	if event.Type != EventTypeStarted {
		t.Errorf("expected type 'task_started', got %s", event.Type)
	}
	if event.TaskID != "test-123" {
		t.Errorf("expected task ID 'test-123', got %s", event.TaskID)
	}
}
