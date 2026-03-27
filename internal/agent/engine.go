package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/llm"
	"github.com/user/deer-flow-go/internal/model"
)

type Engine struct {
	config       *config.Config
	llmFactory   *llm.Factory
	stateManager *StateManager
}

func NewEngine(cfg *config.Config, llmFactory *llm.Factory) *Engine {
	return &Engine{
		config:       cfg,
		llmFactory:   llmFactory,
		stateManager: NewStateManager(),
	}
}

func (e *Engine) Invoke(ctx context.Context, threadID string, input string) (*model.Message, error) {
	state := e.stateManager.GetOrCreate(threadID)

	userMsg := model.Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   input,
		CreatedAt: time.Now(),
	}
	state.Messages = append(state.Messages, userMsg)
	state.UpdatedAt = time.Now()

	if len(e.config.Models) == 0 {
		return nil, fmt.Errorf("no models configured")
	}
	defaultModel := e.config.Models[0].Name

	provider, err := e.llmFactory.Get(defaultModel)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	response, err := provider.Generate(ctx, state.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	response.ID = uuid.New().String()
	response.CreatedAt = time.Now()

	state.Messages = append(state.Messages, *response)
	state.UpdatedAt = time.Now()

	e.stateManager.Update(threadID, state)

	return response, nil
}
