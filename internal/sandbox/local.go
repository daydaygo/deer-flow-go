package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type LocalConfig struct {
	DataDir    string
	SkillsPath string
}

type LocalSandbox struct {
	threadID      string
	dataDir       string
	skillsPath    string
	workspacePath string
	uploadsPath   string
	outputsPath   string
}

type LocalProvider struct {
	config    *LocalConfig
	mu        sync.RWMutex
	sandboxes map[string]*LocalSandbox
}

func NewLocalProvider(cfg *LocalConfig) *LocalProvider {
	return &LocalProvider{
		config:    cfg,
		sandboxes: make(map[string]*LocalSandbox),
	}
}

func (p *LocalProvider) Acquire(ctx context.Context, threadID string) (Sandbox, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if sb, exists := p.sandboxes[threadID]; exists {
		return sb, nil
	}

	threadDir := filepath.Join(p.config.DataDir, "threads", threadID)
	userDataDir := filepath.Join(threadDir, "user-data")

	sb := &LocalSandbox{
		threadID:      threadID,
		dataDir:       p.config.DataDir,
		skillsPath:    p.config.SkillsPath,
		workspacePath: filepath.Join(userDataDir, "workspace"),
		uploadsPath:   filepath.Join(userDataDir, "uploads"),
		outputsPath:   filepath.Join(userDataDir, "outputs"),
	}

	os.MkdirAll(sb.workspacePath, 0755)
	os.MkdirAll(sb.uploadsPath, 0755)
	os.MkdirAll(sb.outputsPath, 0755)

	p.sandboxes[threadID] = sb
	return sb, nil
}

func (p *LocalProvider) Get(threadID string) Sandbox {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if sb, exists := p.sandboxes[threadID]; exists {
		return sb
	}
	return nil
}

func (p *LocalProvider) Release(threadID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sandboxes, threadID)
	return nil
}

func (s *LocalSandbox) translatePath(virtualPath string) string {
	if HasPathTraversal(virtualPath) {
		return virtualPath
	}

	if virtualPath == virtualWorkspacePath || strings.HasPrefix(virtualPath, virtualWorkspacePath+"/") {
		rel := strings.TrimPrefix(virtualPath, virtualWorkspacePath)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return s.workspacePath
		}
		return filepath.Join(s.workspacePath, rel)
	}

	if virtualPath == virtualUploadsPath || strings.HasPrefix(virtualPath, virtualUploadsPath+"/") {
		rel := strings.TrimPrefix(virtualPath, virtualUploadsPath)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return s.uploadsPath
		}
		return filepath.Join(s.uploadsPath, rel)
	}

	if virtualPath == virtualOutputsPath || strings.HasPrefix(virtualPath, virtualOutputsPath+"/") {
		rel := strings.TrimPrefix(virtualPath, virtualOutputsPath)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return s.outputsPath
		}
		return filepath.Join(s.outputsPath, rel)
	}

	if virtualPath == virtualSkillsPath || strings.HasPrefix(virtualPath, virtualSkillsPath+"/") {
		rel := strings.TrimPrefix(virtualPath, virtualSkillsPath)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return s.skillsPath
		}
		return filepath.Join(s.skillsPath, rel)
	}

	return virtualPath
}

func (s *LocalSandbox) reverseTranslatePath(physicalPath string) string {
	physicalPath = filepath.Clean(physicalPath)

	if strings.HasPrefix(physicalPath, s.workspacePath) {
		rel := strings.TrimPrefix(physicalPath, s.workspacePath)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if rel == "" {
			return virtualWorkspacePath
		}
		return virtualWorkspacePath + "/" + rel
	}

	if strings.HasPrefix(physicalPath, s.uploadsPath) {
		rel := strings.TrimPrefix(physicalPath, s.uploadsPath)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if rel == "" {
			return virtualUploadsPath
		}
		return virtualUploadsPath + "/" + rel
	}

	if strings.HasPrefix(physicalPath, s.outputsPath) {
		rel := strings.TrimPrefix(physicalPath, s.outputsPath)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if rel == "" {
			return virtualOutputsPath
		}
		return virtualOutputsPath + "/" + rel
	}

	if strings.HasPrefix(physicalPath, s.skillsPath) {
		rel := strings.TrimPrefix(physicalPath, s.skillsPath)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if rel == "" {
			return virtualSkillsPath
		}
		return virtualSkillsPath + "/" + rel
	}

	return physicalPath
}

func (s *LocalSandbox) maskPathsInOutput(output string) string {
	result := output

	pathsToMask := []struct {
		physical string
		virtual  string
	}{
		{s.skillsPath, virtualSkillsPath},
		{s.workspacePath, virtualWorkspacePath},
		{s.uploadsPath, virtualUploadsPath},
		{s.outputsPath, virtualOutputsPath},
	}

	for _, mapping := range pathsToMask {
		if mapping.physical == "" {
			continue
		}
		escaped := regexp.QuoteMeta(mapping.physical)
		pattern := regexp.MustCompile(escaped + `(?:[^\s\"';&|<>()]*)?`)
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return s.reverseTranslatePath(match)
		})
	}

	return result
}

func (s *LocalSandbox) validatePath(path string, readOnly bool) error {
	if HasPathTraversal(path) {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	if IsSkillsPath(path) {
		if !readOnly {
			return fmt.Errorf("write access to skills path not allowed: %s", path)
		}
		return nil
	}

	if strings.HasPrefix(path, "/mnt/user-data/") {
		return nil
	}

	return fmt.Errorf("path not allowed: %s", path)
}

func (s *LocalSandbox) ExecuteCommand(ctx context.Context, cmd string) (stdout, stderr string, err error) {
	resolvedCmd := s.translatePathsInCommand(cmd)

	shell := "/bin/sh"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "/bin/bash"
	}

	execCmd := exec.CommandContext(ctx, shell, "-c", resolvedCmd)
	var stdoutBuf, stderrBuf bytes.Buffer
	execCmd.Stdout = &stdoutBuf
	execCmd.Stderr = &stderrBuf

	err = execCmd.Run()
	stdout = s.maskPathsInOutput(stdoutBuf.String())
	stderr = s.maskPathsInOutput(stderrBuf.String())

	return stdout, stderr, err
}

func (s *LocalSandbox) translatePathsInCommand(cmd string) string {
	pathsToReplace := []struct {
		virtual  string
		physical string
	}{
		{virtualSkillsPath, s.skillsPath},
		{virtualWorkspacePath, s.workspacePath},
		{virtualUploadsPath, s.uploadsPath},
		{virtualOutputsPath, s.outputsPath},
	}

	result := cmd
	for _, mapping := range pathsToReplace {
		if mapping.physical == "" || mapping.virtual == "" {
			continue
		}
		escaped := regexp.QuoteMeta(mapping.virtual)
		pattern := regexp.MustCompile(escaped + `(?:[^\s\"';&|<>()]*)?`)
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return s.translatePath(match)
		})
	}

	return result
}

func (s *LocalSandbox) ReadFile(ctx context.Context, path string) (string, error) {
	if err := s.validatePath(path, true); err != nil {
		return "", err
	}

	physicalPath := s.translatePath(path)
	data, err := os.ReadFile(physicalPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *LocalSandbox) ReadFileRange(ctx context.Context, path string, startLine, endLine int) (string, error) {
	content, err := s.ReadFile(ctx, path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(content, "\n")
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return "", nil
	}

	selected := lines[startLine-1 : endLine]
	return strings.Join(selected, "\n"), nil
}

func (s *LocalSandbox) WriteFile(ctx context.Context, path, content string) error {
	if err := s.validatePath(path, false); err != nil {
		return err
	}

	physicalPath := s.translatePath(path)
	dir := filepath.Dir(physicalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(physicalPath, []byte(content), 0644)
}

func (s *LocalSandbox) WriteFileAppend(ctx context.Context, path, content string) error {
	if err := s.validatePath(path, false); err != nil {
		return err
	}

	physicalPath := s.translatePath(path)
	dir := filepath.Dir(physicalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(physicalPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func (s *LocalSandbox) ListDir(ctx context.Context, path string) ([]FileInfo, error) {
	if err := s.validatePath(path, true); err != nil {
		return nil, err
	}

	physicalPath := s.translatePath(path)
	entries, err := os.ReadDir(physicalPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		if shouldIgnore(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})

		if entry.IsDir() {
			subPath := filepath.Join(physicalPath, entry.Name())
			subEntries, err := os.ReadDir(subPath)
			if err != nil {
				continue
			}
			for _, subEntry := range subEntries {
				if shouldIgnore(subEntry.Name()) {
					continue
				}
				subInfo, err := subEntry.Info()
				if err != nil {
					continue
				}
				files = append(files, FileInfo{
					Name:  entry.Name() + "/" + subEntry.Name(),
					IsDir: subEntry.IsDir(),
					Size:  subInfo.Size(),
				})
			}
		}
	}

	return files, nil
}

func (s *LocalSandbox) StrReplace(ctx context.Context, path, old, new string, replaceAll bool) error {
	content, err := s.ReadFile(ctx, path)
	if err != nil {
		return err
	}

	if !strings.Contains(content, old) {
		return fmt.Errorf("string to replace not found in file: %s", path)
	}

	if replaceAll {
		content = strings.ReplaceAll(content, old, new)
	} else {
		content = strings.Replace(content, old, new, 1)
	}

	return s.WriteFile(ctx, path, content)
}

func shouldIgnore(name string) bool {
	ignorePatterns := []string{
		".git", ".svn", ".hg", ".bzr",
		"node_modules", "__pycache__", ".venv", "venv", ".env", "env",
		".tox", ".nox", ".eggs", ".egg-info", "site-packages",
		"dist", "build", ".next", ".nuxt", ".output", ".turbo", "target", "out",
		".idea", ".vscode", "*.swp", "*.swo", "*~", ".project", ".classpath", ".settings",
		".DS_Store", "Thumbs.db", "desktop.ini", "*.lnk",
		"*.log", "*.tmp", "*.temp", "*.bak", "*.cache", ".cache", "logs",
		".coverage", "coverage", ".nyc_output", "htmlcov", ".pytest_cache", ".mypy_cache", ".ruff_cache",
	}

	for _, pattern := range ignorePatterns {
		if name == pattern {
			return true
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
	}
	return false
}
