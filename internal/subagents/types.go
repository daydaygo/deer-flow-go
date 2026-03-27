package subagents

import (
	"github.com/user/deer-flow-go/internal/subagentstypes"
)

type Status = subagentstypes.Status

const (
	StatusPending   = subagentstypes.StatusPending
	StatusRunning   = subagentstypes.StatusRunning
	StatusCompleted = subagentstypes.StatusCompleted
	StatusFailed    = subagentstypes.StatusFailed
	StatusTimeout   = subagentstypes.StatusTimeout
)

type Subagent = subagentstypes.Subagent

type Task = subagentstypes.Task

type TaskEvent = subagentstypes.TaskEvent

const (
	EventTypeStarted   = subagentstypes.EventTypeStarted
	EventTypeRunning   = subagentstypes.EventTypeRunning
	EventTypeCompleted = subagentstypes.EventTypeCompleted
	EventTypeFailed    = subagentstypes.EventTypeFailed
	EventTypeTimeout   = subagentstypes.EventTypeTimeout
)

const (
	DefaultMaxTurns       = subagentstypes.DefaultMaxTurns
	DefaultTimeoutSeconds = subagentstypes.DefaultTimeoutSeconds
	MaxConcurrentTasks    = subagentstypes.MaxConcurrentTasks
)

func NewTask(threadID, subagentType, description, prompt string) *Task {
	return subagentstypes.NewTask(threadID, subagentType, description, prompt)
}

func GenerateTaskID() string {
	return subagentstypes.GenerateTaskID()
}
