package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/llm/types"
	"github.com/user/deer-flow-go/internal/model"
)

type OpenAIProvider struct {
	name    string
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewOpenAIProvider(cfg *config.ModelConfig) *OpenAIProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		name:    cfg.Name,
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
}

type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) Generate(ctx context.Context, messages []model.Message, opts ...types.GenerateOption) (*model.Message, error) {
	options := types.ApplyOptions(opts...)

	req := openaiRequest{
		Model:       p.name,
		Messages:    make([]openaiMessage, len(messages)),
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
	}
	if len(options.StopSequences) > 0 {
		req.Stop = options.StopSequences
	}

	for i, m := range messages {
		req.Messages[i] = openaiMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("openai api error: %s", openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]
	return &model.Message{
		Role:      choice.Message.Role,
		Content:   choice.Message.Content,
		CreatedAt: time.Now(),
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, messages []model.Message, opts ...types.GenerateOption) (<-chan types.StreamEvent, error) {
	ch := make(chan types.StreamEvent)
	close(ch)
	return ch, nil
}
