package subagents

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Executor struct {
	registry  *Registry
	tasks     map[string]*Task
	taskMu    sync.RWMutex
	semaphore chan struct{}
	wg        sync.WaitGroup
}

func NewExecutor() *Executor {
	return NewExecutorWithRegistry(globalRegistry)
}

func NewExecutorWithRegistry(registry *Registry) *Executor {
	return &Executor{
		registry:  registry,
		tasks:     make(map[string]*Task),
		semaphore: make(chan struct{}, MaxConcurrentTasks),
	}
}

func (e *Executor) Submit(ctx context.Context, task *Task) (<-chan TaskEvent, error) {
	subagent, err := e.registry.Get(task.SubagentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get subagent: %w", err)
	}

	task.Status = StatusPending
	task.CreatedAt = time.Now()

	e.taskMu.Lock()
	e.tasks[task.TaskID] = task
	e.taskMu.Unlock()

	events := make(chan TaskEvent, 10)

	e.wg.Add(1)
	go e.executeTask(ctx, task, subagent, events)

	return events, nil
}

func (e *Executor) executeTask(ctx context.Context, task *Task, subagent *Subagent, events chan TaskEvent) {
	defer e.wg.Done()
	defer close(events)

	e.semaphore <- struct{}{}
	defer func() { <-e.semaphore }()

	task.Status = StatusRunning
	task.StartedAt = time.Now()

	events <- TaskEvent{
		Type:      EventTypeStarted,
		TaskID:    task.TaskID,
		ThreadID:  task.ThreadID,
		Subagent:  subagent.Name,
		Content:   task.Description,
		Timestamp: time.Now(),
	}

	timeout := time.Duration(subagent.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(DefaultTimeoutSeconds) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := make(chan executeResult, 1)
	go func() {
		res := e.runSubagent(ctx, task, subagent)
		result <- res
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			task.Status = StatusTimeout
			task.Error = fmt.Sprintf("Execution timed out after %s", timeout)
			task.CompletedAt = time.Now()
			events <- TaskEvent{
				Type:      EventTypeTimeout,
				TaskID:    task.TaskID,
				ThreadID:  task.ThreadID,
				Subagent:  subagent.Name,
				Content:   task.Error,
				Timestamp: time.Now(),
			}
		} else {
			task.Status = StatusFailed
			task.Error = ctx.Err().Error()
			task.CompletedAt = time.Now()
			events <- TaskEvent{
				Type:      EventTypeFailed,
				TaskID:    task.TaskID,
				ThreadID:  task.ThreadID,
				Subagent:  subagent.Name,
				Content:   task.Error,
				Timestamp: time.Now(),
			}
		}
	case res := <-result:
		if res.err != nil {
			task.Status = StatusFailed
			task.Error = res.err.Error()
		} else {
			task.Status = StatusCompleted
			task.Result = res.result
		}
		task.AIMessages = res.aiMessages
		task.CompletedAt = time.Now()

		eventType := EventTypeCompleted
		content := task.Result
		if task.Status == StatusFailed {
			eventType = EventTypeFailed
			content = task.Error
		}

		events <- TaskEvent{
			Type:      eventType,
			TaskID:    task.TaskID,
			ThreadID:  task.ThreadID,
			Subagent:  subagent.Name,
			Content:   content,
			Timestamp: time.Now(),
		}
	}
}

type executeResult struct {
	result     string
	err        error
	aiMessages []map[string]any
}

func (e *Executor) runSubagent(ctx context.Context, task *Task, subagent *Subagent) executeResult {
	log.Printf("[subagent] Running subagent %s for task %s", subagent.Name, task.TaskID)

	result := executeResult{
		aiMessages: make([]map[string]any, 0),
	}

	result.result = fmt.Sprintf("Subagent %s executed task: %s", subagent.Name, task.Description)

	return result
}

func (e *Executor) GetTask(taskID string) *Task {
	e.taskMu.RLock()
	defer e.taskMu.RUnlock()
	return e.tasks[taskID]
}

func (e *Executor) ListTasks() []*Task {
	e.taskMu.RLock()
	defer e.taskMu.RUnlock()
	tasks := make([]*Task, 0, len(e.tasks))
	for _, t := range e.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (e *Executor) CleanupTask(taskID string) {
	e.taskMu.Lock()
	defer e.taskMu.Unlock()
	if task, ok := e.tasks[taskID]; ok {
		if task.IsTerminal() {
			delete(e.tasks, taskID)
		}
	}
}

func (e *Executor) Shutdown() {
	e.wg.Wait()
}

var globalExecutor *Executor
var executorOnce sync.Once

func GetExecutor() *Executor {
	executorOnce.Do(func() {
		globalExecutor = NewExecutor()
	})
	return globalExecutor
}
