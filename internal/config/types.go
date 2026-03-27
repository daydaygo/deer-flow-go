package config

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Models  []ModelConfig `mapstructure:"models"`
	Memory  MemoryConfig  `mapstructure:"memory"`
	Storage StorageConfig `mapstructure:"storage"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type ModelConfig struct {
	Name             string `mapstructure:"name"`
	DisplayName      string `mapstructure:"display_name"`
	Use              string `mapstructure:"use"`
	APIKey           string `mapstructure:"api_key"`
	BaseURL          string `mapstructure:"base_url"`
	SupportsThinking bool   `mapstructure:"supports_thinking"`
	SupportsVision   bool   `mapstructure:"supports_vision"`
}

type MemoryConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	StoragePath        string `mapstructure:"storage_path"`
	InjectionEnabled   bool   `mapstructure:"injection_enabled"`
	MaxInjectionTokens int    `mapstructure:"max_injection_tokens"`
}

type StorageConfig struct {
	DataDir string `mapstructure:"data_dir"`
}
