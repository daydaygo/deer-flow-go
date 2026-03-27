package llm

import (
	"testing"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/llm/types"
)

func TestNewFactory(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{
				Name:    "gpt-4",
				Use:     "openai",
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
			},
		},
	}

	f := NewFactory(cfg)
	if f == nil {
		t.Fatal("expected factory to be created")
	}

	if f.cfg != cfg {
		t.Error("expected config to be set")
	}

	if f.provider == nil {
		t.Error("expected provider map to be initialized")
	}
}

func TestFactoryGet(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{
				Name:    "gpt-4",
				Use:     "openai",
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
			},
			{
				Name:    "gpt-3.5-turbo",
				Use:     "openai",
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
			},
		},
	}

	f := NewFactory(cfg)

	t.Run("get existing model", func(t *testing.T) {
		p, err := f.Get("gpt-4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p == nil {
			t.Fatal("expected provider to be returned")
		}
	})

	t.Run("get another existing model", func(t *testing.T) {
		p, err := f.Get("gpt-3.5-turbo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p == nil {
			t.Fatal("expected provider to be returned")
		}
	})

	t.Run("get non-existing model", func(t *testing.T) {
		_, err := f.Get("non-existing-model")
		if err == nil {
			t.Fatal("expected error for non-existing model")
		}
	})

	t.Run("unsupported provider type", func(t *testing.T) {
		cfgWithUnsupported := &config.Config{
			Models: []config.ModelConfig{
				{
					Name: "unsupported-model",
					Use:  "unsupported-provider",
				},
			},
		}
		f := NewFactory(cfgWithUnsupported)
		_, err := f.Get("unsupported-model")
		if err == nil {
			t.Fatal("expected error for unsupported provider")
		}
	})
}

func TestFactoryCaching(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelConfig{
			{
				Name:    "gpt-4",
				Use:     "openai",
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
			},
		},
	}

	f := NewFactory(cfg)

	p1, err := f.Get("gpt-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p2, err := f.Get("gpt-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p1 != p2 {
		t.Error("expected same provider instance from cache")
	}
}

func TestGenerateOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []types.GenerateOption
		expected types.GenerateOptions
	}{
		{
			name: "default options",
			opts: nil,
			expected: types.GenerateOptions{
				Temperature: 0.7,
				MaxTokens:   4096,
			},
		},
		{
			name: "custom temperature",
			opts: []types.GenerateOption{types.WithTemperature(0.5)},
			expected: types.GenerateOptions{
				Temperature: 0.5,
				MaxTokens:   4096,
			},
		},
		{
			name: "custom max tokens",
			opts: []types.GenerateOption{types.WithMaxTokens(2048)},
			expected: types.GenerateOptions{
				Temperature: 0.7,
				MaxTokens:   2048,
			},
		},
		{
			name: "stop sequences",
			opts: []types.GenerateOption{types.WithStopSequences([]string{"END", "STOP"})},
			expected: types.GenerateOptions{
				Temperature:   0.7,
				MaxTokens:     4096,
				StopSequences: []string{"END", "STOP"},
			},
		},
		{
			name: "all options",
			opts: []types.GenerateOption{
				types.WithTemperature(0.3),
				types.WithMaxTokens(1024),
				types.WithStopSequences([]string{"END"}),
			},
			expected: types.GenerateOptions{
				Temperature:   0.3,
				MaxTokens:     1024,
				StopSequences: []string{"END"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.ApplyOptions(tt.opts...)
			if result.Temperature != tt.expected.Temperature {
				t.Errorf("temperature: got %f, want %f", result.Temperature, tt.expected.Temperature)
			}
			if result.MaxTokens != tt.expected.MaxTokens {
				t.Errorf("max tokens: got %d, want %d", result.MaxTokens, tt.expected.MaxTokens)
			}
			if len(result.StopSequences) != len(tt.expected.StopSequences) {
				t.Errorf("stop sequences length: got %d, want %d", len(result.StopSequences), len(tt.expected.StopSequences))
			}
		})
	}
}
