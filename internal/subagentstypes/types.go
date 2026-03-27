package subagentstypes

import (
	"math/rand"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusTimeout   Status = "timeout"
)

type Subagent struct {
	Name            string
	Description     string
	SystemPrompt    string
	Tools           []string
	DisallowedTools []string
	Model           string
	MaxTurns        int
	TimeoutSeconds  int
}

type Task struct {
	TaskID       string
	ThreadID     string
	SubagentType string
	Description  string
	Prompt       string
	MaxTurns     int
	Status       Status
	Result       string
	Error        string
	CreatedAt    time.Time
	StartedAt    time.Time
	CompletedAt  time.Time
	AIMessages   []map[string]any
}

type TaskEvent struct {
	Type      string
	TaskID    string
	ThreadID  string
	Subagent  string
	Content   string
	Timestamp time.Time
}

const (
	EventTypeStarted   = "task_started"
	EventTypeRunning   = "task_running"
	EventTypeCompleted = "task_completed"
	EventTypeFailed    = "task_failed"
	EventTypeTimeout   = "task_timeout"
)

const (
	DefaultMaxTurns       = 50
	DefaultTimeoutSeconds = 900
	MaxConcurrentTasks    = 3
)

func NewTask(threadID, subagentType, description, prompt string) *Task {
	return &Task{
		TaskID:       GenerateTaskID(),
		ThreadID:     threadID,
		SubagentType: subagentType,
		Description:  description,
		Prompt:       prompt,
		MaxTurns:     DefaultMaxTurns,
		Status:       StatusPending,
		CreatedAt:    time.Now(),
		AIMessages:   make([]map[string]any, 0),
	}
}

func GenerateTaskID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (t *Task) IsTerminal() bool {
	return t.Status == StatusCompleted || t.Status == StatusFailed || t.Status == StatusTimeout
}
