package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestSkillsDir(t *testing.T) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")
	publicDir := filepath.Join(skillsDir, "public")
	customDir := filepath.Join(skillsDir, "custom")

	os.MkdirAll(publicDir, 0o755)
	os.MkdirAll(customDir, 0o755)

	skill1Dir := filepath.Join(publicDir, "skill-one")
	os.MkdirAll(skill1Dir, 0o755)
	skill1Content := `---
name: skill-one
description: First test skill
license: MIT
allowed-tools:
  - bash
  - read
---
# Skill One

This is skill one content.
`
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0o644)

	skill2Dir := filepath.Join(customDir, "skill-two")
	os.MkdirAll(skill2Dir, 0o755)
	skill2Content := `---
name: skill-two
description: Second test skill
---
# Skill Two

This is skill two content.
`
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0o644)

	configPath := filepath.Join(tmpDir, "extensions_config.json")
	configContent := `{
  "skills": {
    "skill-one": { "enabled": true }
  }
}`
	os.WriteFile(configPath, []byte(configContent), 0o644)

	return skillsDir, configPath
}

func TestLoader_Load(t *testing.T) {
	skillsDir, configPath := setupTestSkillsDir(t)

	loader := NewLoader(skillsDir, configPath)
	if err := loader.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	skills := loader.List()
	if len(skills) != 2 {
		t.Errorf("len(skills) = %d, want 2", len(skills))
	}

	skill1 := loader.Get("skill-one")
	if skill1 == nil {
		t.Error("Get(skill-one) returned nil")
	} else {
		if skill1.Name != "skill-one" {
			t.Errorf("skill-one.Name = %q, want %q", skill1.Name, "skill-one")
		}
		if skill1.Enabled != true {
			t.Errorf("skill-one.Enabled = %v, want true", skill1.Enabled)
		}
		if len(skill1.AllowedTools) != 2 {
			t.Errorf("skill-one.AllowedTools len = %d, want 2", len(skill1.AllowedTools))
		}
	}

	skill2 := loader.Get("skill-two")
	if skill2 == nil {
		t.Error("Get(skill-two) returned nil")
	} else {
		if skill2.Enabled != false {
			t.Errorf("skill-two.Enabled = %v, want false", skill2.Enabled)
		}
	}
}

func TestLoader_Get_NotFound(t *testing.T) {
	skillsDir, configPath := setupTestSkillsDir(t)

	loader := NewLoader(skillsDir, configPath)
	loader.Load()

	skill := loader.Get("nonexistent")
	if skill != nil {
		t.Errorf("Get(nonexistent) = %v, want nil", skill)
	}
}

func TestLoader_SetEnabled(t *testing.T) {
	skillsDir, configPath := setupTestSkillsDir(t)

	loader := NewLoader(skillsDir, configPath)
	loader.Load()

	if err := loader.SetEnabled("skill-two", true); err != nil {
		t.Fatalf("SetEnabled() error = %v", err)
	}

	skill := loader.Get("skill-two")
	if skill == nil {
		t.Fatal("Get(skill-two) returned nil")
	}
	if skill.Enabled != true {
		t.Errorf("skill-two.Enabled = %v, want true", skill.Enabled)
	}
}

func TestLoader_SetEnabled_NotFound(t *testing.T) {
	skillsDir, configPath := setupTestSkillsDir(t)

	loader := NewLoader(skillsDir, configPath)
	loader.Load()

	err := loader.SetEnabled("nonexistent", true)
	if err == nil {
		t.Error("SetEnabled(nonexistent) expected error")
	}
}

func TestLoader_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")
	os.MkdirAll(filepath.Join(skillsDir, "public"), 0o755)

	configPath := filepath.Join(tmpDir, "extensions_config.json")

	loader := NewLoader(skillsDir, configPath)
	if err := loader.Load(); err != nil {
		t.Fatalf("Load() without config file error = %v", err)
	}

	skills := loader.List()
	if len(skills) != 0 {
		t.Errorf("len(skills) = %d, want 0", len(skills))
	}
}
