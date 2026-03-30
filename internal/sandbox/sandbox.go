package sandbox

import (
	"context"
	"slices"
	"strings"
)

type FileInfo struct {
	Name  string
	IsDir bool
	Size  int64
}

type Sandbox interface {
	ExecuteCommand(ctx context.Context, cmd string) (stdout, stderr string, err error)
	ReadFile(ctx context.Context, path string) (string, error)
	ReadFileRange(ctx context.Context, path string, startLine, endLine int) (string, error)
	WriteFile(ctx context.Context, path, content string) error
	WriteFileAppend(ctx context.Context, path, content string) error
	ListDir(ctx context.Context, path string) ([]FileInfo, error)
	StrReplace(ctx context.Context, path, old, new string, replaceAll bool) error
}

type SandboxProvider interface {
	Acquire(ctx context.Context, threadID string) (Sandbox, error)
	Get(threadID string) Sandbox
	Release(threadID string) error
}

var (
	virtualWorkspacePath = "/mnt/user-data/workspace"
	virtualUploadsPath   = "/mnt/user-data/uploads"
	virtualOutputsPath   = "/mnt/user-data/outputs"
	virtualSkillsPath    = "/mnt/skills"
)

func IsVirtualPath(path string) bool {
	return path == virtualWorkspacePath || strings.HasPrefix(path, virtualWorkspacePath+"/") ||
		path == virtualUploadsPath || strings.HasPrefix(path, virtualUploadsPath+"/") ||
		path == virtualOutputsPath || strings.HasPrefix(path, virtualOutputsPath+"/") ||
		path == virtualSkillsPath || strings.HasPrefix(path, virtualSkillsPath+"/")
}

func IsSkillsPath(path string) bool {
	return path == virtualSkillsPath || strings.HasPrefix(path, virtualSkillsPath+"/")
}

func HasPathTraversal(path string) bool {
	return slices.Contains(strings.Split(path, "/"), "..")
}
