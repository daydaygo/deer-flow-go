package builtins

import (
	"slices"
	"testing"

	"github.com/user/deer-flow-go/internal/subagentstypes"
)

func TestGeneralPurposeSubagent(t *testing.T) {
	s := GeneralPurposeSubagent()

	if s.Name != "general-purpose" {
		t.Errorf("expected name 'general-purpose', got %s", s.Name)
	}
	if s.Description == "" {
		t.Error("expected description to be set")
	}
	if s.SystemPrompt == "" {
		t.Error("expected system prompt to be set")
	}
	if s.Tools != nil {
		t.Error("expected tools to be nil (inherit all)")
	}
	if len(s.DisallowedTools) == 0 {
		t.Error("expected disallowed tools to be set")
	}
	if s.Model != "inherit" {
		t.Errorf("expected model 'inherit', got %s", s.Model)
	}
	if s.MaxTurns <= 0 {
		t.Errorf("expected positive max turns, got %d", s.MaxTurns)
	}
	if s.TimeoutSeconds <= 0 {
		t.Errorf("expected positive timeout, got %d", s.TimeoutSeconds)
	}

	if slices.Contains(s.DisallowedTools, "task") {
		return
	}
	t.Error("expected 'task' to be in disallowed tools to prevent recursion")
}

func TestBashSubagent(t *testing.T) {
	s := BashSubagent()

	if s.Name != "bash" {
		t.Errorf("expected name 'bash', got %s", s.Name)
	}
	if s.Description == "" {
		t.Error("expected description to be set")
	}
	if s.SystemPrompt == "" {
		t.Error("expected system prompt to be set")
	}
	if s.Tools == nil || len(s.Tools) == 0 {
		t.Error("expected tools to be set")
	}
	if len(s.DisallowedTools) == 0 {
		t.Error("expected disallowed tools to be set")
	}
	if s.Model != "inherit" {
		t.Errorf("expected model 'inherit', got %s", s.Model)
	}
	if s.MaxTurns <= 0 {
		t.Errorf("expected positive max turns, got %d", s.MaxTurns)
	}
	if s.TimeoutSeconds <= 0 {
		t.Errorf("expected positive timeout, got %d", s.TimeoutSeconds)
	}

	if slices.Contains(s.DisallowedTools, "task") {
		return
	}
	t.Error("expected 'task' to be in disallowed tools to prevent recursion")
}

func TestBashSubagent_Tools(t *testing.T) {
	s := BashSubagent()

	expectedTools := []string{"bash", "ls", "read_file", "write_file", "str_replace"}
	for _, expected := range expectedTools {
		found := slices.Contains(s.Tools, expected)
		if !found {
			t.Errorf("expected tool '%s' in bash subagent", expected)
		}
	}
}

func TestBuiltinSubagents(t *testing.T) {
	list := BuiltinSubagents()

	if len(list) < 2 {
		t.Errorf("expected at least 2 builtin subagents, got %d", len(list))
	}

	names := make(map[string]bool)
	for _, s := range list {
		names[s.Name] = true
	}

	if !names["general-purpose"] {
		t.Error("expected 'general-purpose' in builtin subagents")
	}
	if !names["bash"] {
		t.Error("expected 'bash' in builtin subagents")
	}
}

func TestSubagentTypes(t *testing.T) {
	general := GeneralPurposeSubagent()
	bash := BashSubagent()

	if general.Name == bash.Name {
		t.Error("expected different names for general-purpose and bash")
	}

	if general.MaxTurns < bash.MaxTurns {
		t.Error("expected general-purpose to have higher max turns")
	}
}

func TestSubagentFields(t *testing.T) {
	s := &subagentstypes.Subagent{
		Name:            "test",
		Description:     "desc",
		SystemPrompt:    "prompt",
		Tools:           []string{"bash"},
		DisallowedTools: []string{"task"},
		Model:           "inherit",
		MaxTurns:        10,
		TimeoutSeconds:  60,
	}

	if s.Name != "test" {
		t.Errorf("expected name 'test', got %s", s.Name)
	}
}
