package skills

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseSkillFile(content string) (*SkillMeta, string, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		return nil, "", fmt.Errorf("skill file must start with YAML frontmatter (---)")
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("invalid skill file format: missing closing ---")
	}

	frontmatter := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	if meta.Name == "" {
		return nil, "", fmt.Errorf("skill name is required in frontmatter")
	}

	return &meta, body, nil
}

func ParseExtensionsConfig(content []byte) (*ExtensionsConfig, error) {
	var config ExtensionsConfig
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	if err := decoder.Decode(&config); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			config.Skills = make(map[string]SkillConfig)
		} else {
			return nil, fmt.Errorf("failed to parse extensions config: %w", err)
		}
	}

	if config.Skills == nil {
		config.Skills = make(map[string]SkillConfig)
	}

	return &config, nil
}
