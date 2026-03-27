package sandbox

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []ToolParameter
}

type ToolParameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

func BashToolDefinition() ToolDefinition {
	return ToolDefinition{
		Name: "bash",
		Description: `Execute a bash command in a Linux environment.

- Use 'python' to run Python code.
- Prefer a thread-local virtual environment in /mnt/user-data/workspace/.venv.
- Use 'python -m pip' (inside the virtual environment) to install Python packages.`,
		Parameters: []ToolParameter{
			{Name: "description", Type: "string", Description: "Explain why you are running this command in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.", Required: true},
			{Name: "command", Type: "string", Description: "The bash command to execute. Always use absolute paths for files and directories.", Required: true},
		},
	}
}

func ReadFileToolDefinition() ToolDefinition {
	return ToolDefinition{
		Name:        "read_file",
		Description: "Read the contents of a text file. Use this to examine source code, configuration files, logs, or any text-based file.",
		Parameters: []ToolParameter{
			{Name: "description", Type: "string", Description: "Explain why you are reading this file in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.", Required: true},
			{Name: "path", Type: "string", Description: "The absolute path to the file to read.", Required: true},
			{Name: "start_line", Type: "integer", Description: "Optional starting line number (1-indexed, inclusive). Use with end_line to read a specific range.", Required: false},
			{Name: "end_line", Type: "integer", Description: "Optional ending line number (1-indexed, inclusive). Use with start_line to read a specific range.", Required: false},
		},
	}
}

func WriteFileToolDefinition() ToolDefinition {
	return ToolDefinition{
		Name:        "write_file",
		Description: "Write text content to a file.",
		Parameters: []ToolParameter{
			{Name: "description", Type: "string", Description: "Explain why you are writing to this file in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.", Required: true},
			{Name: "path", Type: "string", Description: "The absolute path to the file to write to. ALWAYS PROVIDE THIS PARAMETER SECOND.", Required: true},
			{Name: "content", Type: "string", Description: "The content to write to the file. ALWAYS PROVIDE THIS PARAMETER THIRD.", Required: true},
			{Name: "append", Type: "boolean", Description: "Whether to append the content to the file. If false, the file will be created or overwritten.", Required: false},
		},
	}
}

func StrReplaceToolDefinition() ToolDefinition {
	return ToolDefinition{
		Name: "str_replace",
		Description: `Replace a substring in a file with another substring.
If 'replace_all' is false (default), the substring to replace must appear exactly once in the file.`,
		Parameters: []ToolParameter{
			{Name: "description", Type: "string", Description: "Explain why you are replacing the substring in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.", Required: true},
			{Name: "path", Type: "string", Description: "The absolute path to the file to replace the substring in. ALWAYS PROVIDE THIS PARAMETER SECOND.", Required: true},
			{Name: "old_str", Type: "string", Description: "The substring to replace. ALWAYS PROVIDE THIS PARAMETER THIRD.", Required: true},
			{Name: "new_str", Type: "string", Description: "The new substring. ALWAYS PROVIDE THIS PARAMETER FOURTH.", Required: true},
			{Name: "replace_all", Type: "boolean", Description: "Whether to replace all occurrences of the substring. If false, only the first occurrence will be replaced. Default is false.", Required: false},
		},
	}
}

func LsToolDefinition() ToolDefinition {
	return ToolDefinition{
		Name:        "ls",
		Description: "List the contents of a directory up to 2 levels deep in tree format.",
		Parameters: []ToolParameter{
			{Name: "description", Type: "string", Description: "Explain why you are listing this directory in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.", Required: true},
			{Name: "path", Type: "string", Description: "The absolute path to the directory to list.", Required: true},
		},
	}
}

func AllToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		BashToolDefinition(),
		ReadFileToolDefinition(),
		WriteFileToolDefinition(),
		StrReplaceToolDefinition(),
		LsToolDefinition(),
	}
}
