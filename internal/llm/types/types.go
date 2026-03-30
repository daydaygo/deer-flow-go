package types

import (
	"context"

	"github.com/user/deer-flow-go/internal/model"
)

type GenerateOption func(*GenerateOptions)

type GenerateOptions struct {
	Temperature   float64
	MaxTokens     int
	StopSequences []string
}

func WithTemperature(t float64) GenerateOption {
	return func(o *GenerateOptions) {
		o.Temperature = t
	}
}

func WithMaxTokens(n int) GenerateOption {
	return func(o *GenerateOptions) {
		o.MaxTokens = n
	}
}

func WithStopSequences(seqs []string) GenerateOption {
	return func(o *GenerateOptions) {
		o.StopSequences = seqs
	}
}

func ApplyOptions(opts ...GenerateOption) *GenerateOptions {
	o := &GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   4096,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type StreamEvent struct {
	Type    string
	Content string
	Data    any
}

type Provider interface {
	Generate(ctx context.Context, messages []model.Message, opts ...GenerateOption) (*model.Message, error)
	Stream(ctx context.Context, messages []model.Message, opts ...GenerateOption) (<-chan StreamEvent, error)
}
