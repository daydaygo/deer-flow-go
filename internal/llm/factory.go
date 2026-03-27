package llm

import (
	"fmt"
	"sync"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/llm/providers"
	"github.com/user/deer-flow-go/internal/llm/types"
)

type Factory struct {
	cfg      *config.Config
	provider map[string]types.Provider
	mu       sync.RWMutex
}

func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		cfg:      cfg,
		provider: make(map[string]types.Provider),
	}
}

func (f *Factory) Get(name string) (types.Provider, error) {
	f.mu.RLock()
	if p, ok := f.provider[name]; ok {
		f.mu.RUnlock()
		return p, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if p, ok := f.provider[name]; ok {
		return p, nil
	}

	modelCfg := f.findModelConfig(name)
	if modelCfg == nil {
		return nil, fmt.Errorf("model %q not found in configuration", name)
	}

	p, err := f.createProvider(modelCfg)
	if err != nil {
		return nil, err
	}

	f.provider[name] = p
	return p, nil
}

func (f *Factory) findModelConfig(name string) *config.ModelConfig {
	for i := range f.cfg.Models {
		if f.cfg.Models[i].Name == name {
			return &f.cfg.Models[i]
		}
	}
	return nil
}

func (f *Factory) createProvider(cfg *config.ModelConfig) (types.Provider, error) {
	switch cfg.Use {
	case "openai":
		return providers.NewOpenAIProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Use)
	}
}
