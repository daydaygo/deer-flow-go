package subagents

import (
	"testing"

	"github.com/user/deer-flow-go/internal/subagents/builtins"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected registry to be created")
	}
	if r.subagents == nil {
		t.Error("expected subagents map to be initialized")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	s := &Subagent{
		Name:         "test-agent",
		Description:  "Test agent",
		SystemPrompt: "Test prompt",
	}

	r.Register(s)

	retrieved, err := r.Get("test-agent")
	if err != nil {
		t.Fatalf("expected to retrieve registered subagent: %v", err)
	}
	if retrieved.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %s", retrieved.Name)
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("non-existent")
	if err == nil {
		t.Error("expected error for non-existent subagent")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.RegisterBuiltins(builtins.BuiltinSubagents())

	list := r.List()
	if len(list) < 2 {
		t.Errorf("expected at least 2 builtin subagents, got %d", len(list))
	}

	names := make(map[string]bool)
	for _, s := range list {
		names[s.Name] = true
	}

	if !names["general-purpose"] {
		t.Error("expected 'general-purpose' subagent to be registered")
	}
	if !names["bash"] {
		t.Error("expected 'bash' subagent to be registered")
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	r.RegisterBuiltins(builtins.BuiltinSubagents())

	names := r.Names()
	if len(names) < 2 {
		t.Errorf("expected at least 2 names, got %d", len(names))
	}
}

func TestGlobalRegistry(t *testing.T) {
	s, err := Get("general-purpose")
	if err != nil {
		t.Fatalf("expected to get 'general-purpose' from global registry: %v", err)
	}
	if s.Name != "general-purpose" {
		t.Errorf("expected name 'general-purpose', got %s", s.Name)
	}

	list := List()
	if len(list) < 2 {
		t.Errorf("expected at least 2 subagents in global list, got %d", len(list))
	}

	names := Names()
	if len(names) < 2 {
		t.Errorf("expected at least 2 names in global list, got %d", len(names))
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	r := NewRegistry()
	done := make(chan bool)

	for i := range 100 {
		go func(id int) {
			name := "agent-" + string(rune(id))
			r.Register(&Subagent{Name: name})
			r.Get(name)
			r.List()
			done <- true
		}(i)
	}

	for range 100 {
		<-done
	}
}
