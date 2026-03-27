package subagents

import (
	"github.com/user/deer-flow-go/internal/subagents/builtins"
)

func init() {
	InitBuiltins(builtins.BuiltinSubagents())
}
