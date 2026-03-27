package skills

import (
	"testing"
)

func TestParseSkillFile_Valid(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
license: MIT
allowed-tools:
  - tool1
  - tool2
---
# Test Skill Content

This is the skill content.
`

	meta, body, err := ParseSkillFile(content)
	if err != nil {
		t.Fatalf("ParseSkillFile() error = %v", err)
	}

	if meta.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", meta.Name, "test-skill")
	}
	if meta.Description != "A test skill" {
		t.Errorf("Description = %q, want %q", meta.Description, "A test skill")
	}
	if meta.License != "MIT" {
		t.Errorf("License = %q, want %q", meta.License, "MIT")
	}
	if len(meta.AllowedTools) != 2 {
		t.Errorf("len(AllowedTools) = %d, want 2", len(meta.AllowedTools))
	}
	if meta.AllowedTools[0] != "tool1" {
		t.Errorf("AllowedTools[0] = %q, want %q", meta.AllowedTools[0], "tool1")
	}
	if body != "# Test Skill Content\n\nThis is the skill content." {
		t.Errorf("Body = %q, want %q", body, "# Test Skill Content\n\nThis is the skill content.")
	}
}

func TestParseSkillFile_NoFrontmatter(t *testing.T) {
	content := "Just content without frontmatter"

	_, _, err := ParseSkillFile(content)
	if err == nil {
		t.Error("ParseSkillFile() expected error for missing frontmatter")
	}
}

func TestParseSkillFile_MissingClosingDelimiter(t *testing.T) {
	content := `---
name: test-skill
`

	_, _, err := ParseSkillFile(content)
	if err == nil {
		t.Error("ParseSkillFile() expected error for missing closing delimiter")
	}
}

func TestParseSkillFile_MissingName(t *testing.T) {
	content := `---
description: A test skill
---
Content
`

	_, _, err := ParseSkillFile(content)
	if err == nil {
		t.Error("ParseSkillFile() expected error for missing name")
	}
}

func TestParseSkillFile_EmptyAllowedTools(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
---
Content
`

	meta, _, err := ParseSkillFile(content)
	if err != nil {
		t.Fatalf("ParseSkillFile() error = %v", err)
	}

	if len(meta.AllowedTools) != 0 {
		t.Errorf("len(AllowedTools) = %d, want 0", len(meta.AllowedTools))
	}
}

func TestParseExtensionsConfig_Valid(t *testing.T) {
	content := []byte(`{
  "skills": {
    "example-skill": { "enabled": true },
    "another-skill": { "enabled": false }
  }
}`)

	config, err := ParseExtensionsConfig(content)
	if err != nil {
		t.Fatalf("ParseExtensionsConfig() error = %v", err)
	}

	if len(config.Skills) != 2 {
		t.Errorf("len(Skills) = %d, want 2", len(config.Skills))
	}

	if config.Skills["example-skill"].Enabled != true {
		t.Errorf("example-skill.Enabled = %v, want true", config.Skills["example-skill"].Enabled)
	}

	if config.Skills["another-skill"].Enabled != false {
		t.Errorf("another-skill.Enabled = %v, want false", config.Skills["another-skill"].Enabled)
	}
}

func TestParseExtensionsConfig_Empty(t *testing.T) {
	content := []byte("{}")

	config, err := ParseExtensionsConfig(content)
	if err != nil {
		t.Fatalf("ParseExtensionsConfig() error = %v", err)
	}

	if len(config.Skills) != 0 {
		t.Errorf("len(Skills) = %d, want 0", len(config.Skills))
	}
}
