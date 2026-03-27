package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	once         sync.Once
	mu           sync.RWMutex
)

func Load(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	dir := filepath.Dir(absPath)
	envPath := filepath.Join(dir, ".env")
	_ = godotenv.Load(envPath)

	v := viper.New()
	v.SetConfigFile(absPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	resolveEnvVars(&cfg)

	mu.Lock()
	globalConfig = &cfg
	mu.Unlock()

	return &cfg, nil
}

func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return globalConfig
}

func resolveEnvVars(cfg *Config) {
	for i := range cfg.Models {
		cfg.Models[i].APIKey = resolveEnvVar(cfg.Models[i].APIKey)
		cfg.Models[i].BaseURL = resolveEnvVar(cfg.Models[i].BaseURL)
	}
}

func resolveEnvVar(value string) string {
	if strings.HasPrefix(value, "$") {
		envName := strings.TrimPrefix(value, "$")
		return os.Getenv(envName)
	}
	return value
}
