package subagents

import (
	"context"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	e := NewExecutor()
	if e == nil {
		t.Fatal("expected executor to be created")
	}
	if e.registry == nil {
		t.Error("expected registry to be initialized")
	}
	if e.tasks == nil {
		t.Error("expected tasks map to be initialized")
	}
}

func TestExecutor_Submit_InvalidSubagent(t *testing.T) {
	e := NewExecutor()
	task := &Task{
		TaskID:       "test-1",
		ThreadID:     "thread-1",
		SubagentType: "non-existent",
		Description:  "test",
		Prompt:       "test prompt",
	}

	_, err := e.Submit(context.Background(), task)
	if err == nil {
		t.Error("expected error for non-existent subagent")
	}
}

func TestExecutor_Submit_ValidTask(t *testing.T) {
	e := NewExecutor()
	task := NewTask("thread-1", "general-purpose", "test task", "test prompt")

	events, err := e.Submit(context.Background(), task)
	if err != nil {
		t.Fatalf("expected successful submit: %v", err)
	}
	if events == nil {
		t.Fatal("expected events channel")
	}

	retrieved := e.GetTask(task.TaskID)
	if retrieved == nil {
		t.Fatal("expected task to be stored")
	}
	if retrieved.TaskID != task.TaskID {
		t.Errorf("expected task ID %s, got %s", task.TaskID, retrieved.TaskID)
	}
}

func TestExecutor_GetTask(t *testing.T) {
	e := NewExecutor()

	task := e.GetTask("non-existent")
	if task != nil {
		t.Error("expected nil for non-existent task")
	}

	newTask := NewTask("thread-1", "general-purpose", "test", "test")
	e.Submit(context.Background(), newTask)

	retrieved := e.GetTask(newTask.TaskID)
	if retrieved == nil {
		t.Fatal("expected to retrieve task")
	}
}

func TestExecutor_ListTasks(t *testing.T) {
	e := NewExecutor()

	tasks := e.ListTasks()
	if len(tasks) != 0 {
		t.Errorf("expected empty task list, got %d", len(tasks))
	}

	task1 := &Task{
		TaskID:       "task-1-unique",
		ThreadID:     "thread-1",
		SubagentType: "general-purpose",
		Description:  "task1",
		Prompt:       "prompt1",
		Status:       StatusPending,
		CreatedAt:    time.Now(),
		AIMessages:   make([]map[string]any, 0),
	}
	task2 := &Task{
		TaskID:       "task-2-unique",
		ThreadID:     "thread-2",
		SubagentType: "bash",
		Description:  "task2",
		Prompt:       "prompt2",
		Status:       StatusPending,
		CreatedAt:    time.Now(),
		AIMessages:   make([]map[string]any, 0),
	}

	e.Submit(context.Background(), task1)
	e.Submit(context.Background(), task2)

	tasks = e.ListTasks()
	if len(tasks) < 2 {
		t.Errorf("expected at least 2 tasks, got %d", len(tasks))
	}
}

func TestExecutor_CleanupTask(t *testing.T) {
	e := NewExecutor()
	task := NewTask("thread-1", "general-purpose", "test", "test")
	e.Submit(context.Background(), task)

	task.Status = StatusCompleted
	task.CompletedAt = time.Now()

	e.CleanupTask(task.TaskID)

	retrieved := e.GetTask(task.TaskID)
	if retrieved != nil {
		t.Error("expected task to be cleaned up")
	}
}

func TestExecutor_CleanupTask_NonTerminal(t *testing.T) {
	e := NewExecutor()
	task := NewTask("thread-1", "general-purpose", "test", "test")
	e.Submit(context.Background(), task)

	task.Status = StatusRunning

	e.CleanupTask(task.TaskID)

	retrieved := e.GetTask(task.TaskID)
	if retrieved == nil {
		t.Error("expected task not to be cleaned up for non-terminal status")
	}
}

func TestExecutor_Events(t *testing.T) {
	e := NewExecutor()
	task := NewTask("thread-1", "general-purpose", "test task", "test prompt")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := e.Submit(ctx, task)
	if err != nil {
		t.Fatalf("expected successful submit: %v", err)
	}

	eventCount := 0
	for event := range events {
		eventCount++
		if event.TaskID != task.TaskID {
			t.Errorf("expected task ID %s, got %s", task.TaskID, event.TaskID)
		}
		if event.Timestamp.IsZero() {
			t.Error("expected timestamp to be set")
		}
	}

	if eventCount < 2 {
		t.Errorf("expected at least 2 events, got %d", eventCount)
	}
}

func TestExecutor_Timeout(t *testing.T) {
	e := NewExecutor()
	task := NewTask("thread-1", "general-purpose", "test task", "test prompt")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	events, err := e.Submit(ctx, task)
	if err != nil {
		t.Fatalf("expected successful submit: %v", err)
	}

	for range events {
	}

	retrieved := e.GetTask(task.TaskID)
	if retrieved == nil {
		t.Fatal("expected task to be stored")
	}

	if retrieved.Status != StatusTimeout && retrieved.Status != StatusFailed {
		t.Errorf("expected timeout or failed status, got %s", retrieved.Status)
	}
}

func TestExecutor_Shutdown(t *testing.T) {
	e := NewExecutor()
	task := NewTask("thread-1", "general-purpose", "test", "test")

	e.Submit(context.Background(), task)

	e.Shutdown()
}

func TestExecutor_ConcurrencyLimit(t *testing.T) {
	e := NewExecutor()

	if cap(e.semaphore) != MaxConcurrentTasks {
		t.Errorf("expected semaphore capacity %d, got %d", MaxConcurrentTasks, cap(e.semaphore))
	}
}
