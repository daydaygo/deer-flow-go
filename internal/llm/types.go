package llm

import (
	"github.com/user/deer-flow-go/internal/llm/types"
)

type (
	GenerateOption  = types.GenerateOption
	GenerateOptions = types.GenerateOptions
	StreamEvent     = types.StreamEvent
	Provider        = types.Provider
)

var (
	WithTemperature   = types.WithTemperature
	WithMaxTokens     = types.WithMaxTokens
	WithStopSequences = types.WithStopSequences
	ApplyOptions      = types.ApplyOptions
)
