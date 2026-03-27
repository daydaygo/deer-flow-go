package agent

import (
	"context"
	"testing"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/llm"
	"github.com/user/deer-flow-go/internal/model"
)

func TestStateManager_Get(t *testing.T) {
	sm := NewStateManager()

	state := sm.Get("non-existent")
	if state != nil {
		t.Error("expected nil for non-existent thread")
	}

	created := sm.GetOrCreate("thread-1")
	if created == nil {
		t.Fatal("expected state to be created")
	}
	if created.ThreadID != "thread-1" {
		t.Errorf("expected thread ID 'thread-1', got %s", created.ThreadID)
	}

	retrieved := sm.Get("thread-1")
	if retrieved == nil {
		t.Fatal("expected to retrieve state")
	}
	if retrieved.ThreadID != "thread-1" {
		t.Errorf("expected thread ID 'thread-1', got %s", retrieved.ThreadID)
	}
}

func TestStateManager_GetOrCreate(t *testing.T) {
	sm := NewStateManager()

	state1 := sm.GetOrCreate("thread-1")
	if state1 == nil {
		t.Fatal("expected state to be created")
	}
	if state1.ThreadID != "thread-1" {
		t.Errorf("expected thread ID 'thread-1', got %s", state1.ThreadID)
	}
	if state1.Messages == nil {
		t.Error("expected messages slice to be initialized")
	}

	state2 := sm.GetOrCreate("thread-1")
	if state1 != state2 {
		t.Error("expected same state instance for same thread ID")
	}
}

func TestStateManager_Update(t *testing.T) {
	sm := NewStateManager()

	state := sm.GetOrCreate("thread-1")
	state.Title = "Test Thread"

	sm.Update("thread-1", state)

	updated := sm.Get("thread-1")
	if updated == nil {
		t.Fatal("expected to retrieve updated state")
	}
	if updated.Title != "Test Thread" {
		t.Errorf("expected title 'Test Thread', got %s", updated.Title)
	}
}

type mockProvider struct {
	response *model.Message
	err      error
}

func (m *mockProvider) Generate(ctx context.Context, messages []model.Message, opts ...llm.GenerateOption) (*model.Message, error) {
	return m.response, m.err
}

func (m *mockProvider) Stream(ctx context.Context, messages []model.Message, opts ...llm.GenerateOption) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

func TestNewEngine(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{Name: "test-model", Use: "openai", APIKey: "test-key"},
		},
	}
	factory := llm.NewFactory(cfg)

	engine := NewEngine(cfg, factory)
	if engine == nil {
		t.Fatal("expected engine to be created")
	}
	if engine.stateManager == nil {
		t.Error("expected state manager to be initialized")
	}
}

func TestEngine_Invoke_NoModels(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{},
	}
	factory := llm.NewFactory(cfg)
	engine := NewEngine(cfg, factory)

	_, err := engine.Invoke(context.Background(), "thread-1", "hello")
	if err == nil {
		t.Error("expected error when no models configured")
	}
}

func TestEngine_Invoke_InvalidModel(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{Name: "gpt-4o", Use: "invalid-provider", APIKey: "test-key"},
		},
	}
	factory := llm.NewFactory(cfg)
	engine := NewEngine(cfg, factory)

	_, err := engine.Invoke(context.Background(), "thread-1", "hello")
	if err == nil {
		t.Error("expected error for invalid provider")
	}
}

func TestStateManager_Concurrency(t *testing.T) {
	sm := NewStateManager()
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(id int) {
			threadID := "thread-" + string(rune(id))
			sm.GetOrCreate(threadID)
			sm.Update(threadID, &model.ThreadState{ThreadID: threadID})
			sm.Get(threadID)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
