package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Loader struct {
	skillsPath       string
	configPath       string
	skills           []Skill
	skillsByName     map[string]*Skill
	extensionsConfig *ExtensionsConfig
	mu               sync.RWMutex
}

func NewLoader(skillsPath, configPath string) *Loader {
	return &Loader{
		skillsPath:   skillsPath,
		configPath:   configPath,
		skillsByName: make(map[string]*Skill),
	}
}

func (l *Loader) Load() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.skills = nil
	l.skillsByName = make(map[string]*Skill)

	if err := l.loadExtensionsConfig(); err != nil {
		return fmt.Errorf("failed to load extensions config: %w", err)
	}

	dirs := []string{"public", "custom"}
	for _, dir := range dirs {
		dirPath := filepath.Join(l.skillsPath, dir)
		if err := l.loadSkillsFromDir(dirPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to load skills from %s: %w", dir, err)
		}
	}

	return nil
}

func (l *Loader) loadExtensionsConfig() error {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			l.extensionsConfig = &ExtensionsConfig{Skills: make(map[string]SkillConfig)}
			return nil
		}
		return err
	}

	config, err := ParseExtensionsConfig(data)
	if err != nil {
		return err
	}

	l.extensionsConfig = config
	return nil
}

func (l *Loader) loadSkillsFromDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(dirPath, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")

		content, err := os.ReadFile(skillFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read skill file %s: %w", skillFile, err)
		}

		meta, body, err := ParseSkillFile(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse skill file %s: %w", skillFile, err)
		}

		skill := Skill{
			Name:         meta.Name,
			Description:  meta.Description,
			License:      meta.License,
			AllowedTools: meta.AllowedTools,
			Path:         skillDir,
			Content:      body,
			Enabled:      l.isSkillEnabled(meta.Name),
		}

		l.skills = append(l.skills, skill)
		l.skillsByName[skill.Name] = &l.skills[len(l.skills)-1]
	}

	return nil
}

func (l *Loader) isSkillEnabled(name string) bool {
	if l.extensionsConfig == nil {
		return false
	}

	config, exists := l.extensionsConfig.Skills[name]
	if !exists {
		return false
	}

	return config.Enabled
}

func (l *Loader) List() []Skill {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]Skill, len(l.skills))
	copy(result, l.skills)
	return result
}

func (l *Loader) Get(name string) *Skill {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if skill, exists := l.skillsByName[name]; exists {
		skillCopy := *skill
		return &skillCopy
	}
	return nil
}

func (l *Loader) SetEnabled(name string, enabled bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	skill, exists := l.skillsByName[name]
	if !exists {
		return fmt.Errorf("skill not found: %s", name)
	}

	skill.Enabled = enabled

	if l.extensionsConfig.Skills == nil {
		l.extensionsConfig.Skills = make(map[string]SkillConfig)
	}
	l.extensionsConfig.Skills[name] = SkillConfig{Enabled: enabled}

	if err := l.saveExtensionsConfig(); err != nil {
		return err
	}

	return nil
}

func (l *Loader) saveExtensionsConfig() error {
	data, err := json.MarshalIndent(l.extensionsConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal extensions config: %w", err)
	}

	if err := os.WriteFile(l.configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write extensions config: %w", err)
	}

	return nil
}
