package builtins

import "github.com/user/deer-flow-go/internal/subagentstypes"

func BuiltinSubagents() []*subagentstypes.Subagent {
	return []*subagentstypes.Subagent{
		GeneralPurposeSubagent(),
		BashSubagent(),
	}
}
