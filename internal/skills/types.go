package skills

type Skill struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	License      string   `json:"license,omitempty"`
	AllowedTools []string `json:"allowed_tools,omitempty"`
	Path         string   `json:"path"`
	Content      string   `json:"content,omitempty"`
	Enabled      bool     `json:"enabled"`
}

type SkillMeta struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	License      string   `yaml:"license"`
	AllowedTools []string `yaml:"allowed-tools"`
}

type SkillConfig struct {
	Enabled bool `json:"enabled"`
}

type ExtensionsConfig struct {
	Skills map[string]SkillConfig `json:"skills"`
}
