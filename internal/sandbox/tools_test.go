package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalSandbox_PathTranslation(t *testing.T) {
	tmpDir := t.TempDir()
	skillsPath := filepath.Join(tmpDir, "skills")
	os.MkdirAll(skillsPath, 0755)

	threadID := "test-thread-123"
	dataDir := filepath.Join(tmpDir, "data")

	cfg := &LocalConfig{
		DataDir:    dataDir,
		SkillsPath: skillsPath,
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	tests := []struct {
		virtual      string
		wantPhysical string
	}{
		{
			virtual:      "/mnt/user-data/workspace",
			wantPhysical: filepath.Join(dataDir, "threads", threadID, "user-data", "workspace"),
		},
		{
			virtual:      "/mnt/user-data/workspace/foo.txt",
			wantPhysical: filepath.Join(dataDir, "threads", threadID, "user-data", "workspace", "foo.txt"),
		},
		{
			virtual:      "/mnt/user-data/uploads",
			wantPhysical: filepath.Join(dataDir, "threads", threadID, "user-data", "uploads"),
		},
		{
			virtual:      "/mnt/user-data/outputs",
			wantPhysical: filepath.Join(dataDir, "threads", threadID, "user-data", "outputs"),
		},
		{
			virtual:      "/mnt/skills",
			wantPhysical: skillsPath,
		},
		{
			virtual:      "/mnt/skills/public/SKILL.md",
			wantPhysical: filepath.Join(skillsPath, "public", "SKILL.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.virtual, func(t *testing.T) {
			ls := sb.(*LocalSandbox)
			got := ls.translatePath(tt.virtual)
			if got != tt.wantPhysical {
				t.Errorf("translatePath(%q) = %q, want %q", tt.virtual, got, tt.wantPhysical)
			}
		})
	}
}

func TestLocalSandbox_ExecuteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	threadID := "test-thread-123"

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()
	stdout, stderr, err := sb.ExecuteCommand(ctx, "echo hello")
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}
	if !strings.Contains(stdout, "hello") {
		t.Errorf("stdout should contain 'hello', got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}
}

func TestLocalSandbox_WriteAndReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	threadID := "test-thread-123"

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()

	testContent := "Hello, World!\nLine 2\nLine 3"
	virtualPath := "/mnt/user-data/workspace/test.txt"

	err = sb.WriteFile(ctx, virtualPath, testContent)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := sb.ReadFile(ctx, virtualPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if got != testContent {
		t.Errorf("ReadFile content mismatch: got %q, want %q", got, testContent)
	}
}

func TestLocalSandbox_ReadFileWithLineRange(t *testing.T) {
	tmpDir := t.TempDir()
	threadID := "test-thread-123"

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()

	content := "Line1\nLine2\nLine3\nLine4\nLine5"
	virtualPath := "/mnt/user-data/workspace/test.txt"

	err = sb.WriteFile(ctx, virtualPath, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := sb.ReadFileRange(ctx, virtualPath, 2, 4)
	if err != nil {
		t.Fatalf("ReadFileRange failed: %v", err)
	}
	want := "Line2\nLine3\nLine4"
	if got != want {
		t.Errorf("ReadFileRange(2,4) = %q, want %q", got, want)
	}
}

func TestLocalSandbox_ListDir(t *testing.T) {
	tmpDir := t.TempDir()
	threadID := "test-thread-123"

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()

	workspacePath := "/mnt/user-data/workspace"
	sb.WriteFile(ctx, workspacePath+"/file1.txt", "content1")
	sb.WriteFile(ctx, workspacePath+"/file2.txt", "content2")
	sb.WriteFile(ctx, workspacePath+"/subdir/file3.txt", "content3")

	files, err := sb.ListDir(ctx, workspacePath)
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}

	if len(files) == 0 {
		t.Error("ListDir returned empty result")
	}

	var foundFile1, foundFile2, foundSubdir bool
	for _, f := range files {
		if f.Name == "file1.txt" && !f.IsDir {
			foundFile1 = true
		}
		if f.Name == "file2.txt" && !f.IsDir {
			foundFile2 = true
		}
		if f.Name == "subdir" && f.IsDir {
			foundSubdir = true
		}
	}

	if !foundFile1 || !foundFile2 || !foundSubdir {
		t.Errorf("ListDir missing expected files: foundFile1=%v, foundFile2=%v, foundSubdir=%v",
			foundFile1, foundFile2, foundSubdir)
	}
}

func TestLocalSandbox_StrReplace(t *testing.T) {
	tmpDir := t.TempDir()
	threadID := "test-thread-123"

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()

	content := "Hello World\nHello World\nGoodbye"
	virtualPath := "/mnt/user-data/workspace/test.txt"

	err = sb.WriteFile(ctx, virtualPath, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = sb.StrReplace(ctx, virtualPath, "Hello", "Hi", false)
	if err != nil {
		t.Fatalf("StrReplace failed: %v", err)
	}

	got, err := sb.ReadFile(ctx, virtualPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	want := "Hi World\nHello World\nGoodbye"
	if got != want {
		t.Errorf("StrReplace single: got %q, want %q", got, want)
	}

	err = sb.StrReplace(ctx, virtualPath, "Hello", "Hey", true)
	if err != nil {
		t.Fatalf("StrReplace replaceAll failed: %v", err)
	}

	got, err = sb.ReadFile(ctx, virtualPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	want = "Hi World\nHey World\nGoodbye"
	if got != want {
		t.Errorf("StrReplace replaceAll: got %q, want %q", got, want)
	}
}

func TestLocalSandbox_PathTraversalRejection(t *testing.T) {
	tmpDir := t.TempDir()
	threadID := "test-thread-123"

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()

	tests := []string{
		"/mnt/user-data/workspace/../uploads/sneak.txt",
		"/mnt/user-data/workspace/../../data/sneak.txt",
		"/mnt/skills/../sneak.txt",
	}

	for _, path := range tests {
		err := sb.WriteFile(ctx, path, "test")
		if err == nil {
			t.Errorf("WriteFile(%q) should reject path traversal", path)
		}
	}
}

func TestLocalSandbox_SkillsPathReadOnly(t *testing.T) {
	tmpDir := t.TempDir()
	skillsPath := filepath.Join(tmpDir, "skills")
	os.MkdirAll(filepath.Join(skillsPath, "public"), 0755)
	os.WriteFile(filepath.Join(skillsPath, "public", "SKILL.md"), []byte("skill content"), 0644)

	threadID := "test-thread-123"
	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: skillsPath,
	}

	provider := NewLocalProvider(cfg)
	sb, err := provider.Acquire(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	ctx := context.Background()

	_, err = sb.ReadFile(ctx, "/mnt/skills/public/SKILL.md")
	if err != nil {
		t.Errorf("ReadFile from skills path should succeed: %v", err)
	}

	err = sb.WriteFile(ctx, "/mnt/skills/public/attack.txt", "malicious")
	if err == nil {
		t.Error("WriteFile to skills path should be rejected")
	}
}

func TestLocalProvider_GetAndRelease(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &LocalConfig{
		DataDir:    filepath.Join(tmpDir, "data"),
		SkillsPath: filepath.Join(tmpDir, "skills"),
	}

	provider := NewLocalProvider(cfg)

	threadID := "thread-1"
	ctx := context.Background()

	sb1, err := provider.Acquire(ctx, threadID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	sb2 := provider.Get(threadID)
	if sb2 == nil {
		t.Error("Get should return the sandbox")
	}
	if sb1 != sb2 {
		t.Error("Get should return same sandbox instance")
	}

	err = provider.Release(threadID)
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	sb3 := provider.Get(threadID)
	if sb3 != nil {
		t.Error("Get after Release should return nil")
	}
}

func TestBashToolDefinition(t *testing.T) {
	tool := BashToolDefinition()
	if tool.Name != "bash" {
		t.Errorf("tool.Name = %q, want 'bash'", tool.Name)
	}
	if tool.Description == "" {
		t.Error("tool.Description should not be empty")
	}
}

func TestReadFileToolDefinition(t *testing.T) {
	tool := ReadFileToolDefinition()
	if tool.Name != "read_file" {
		t.Errorf("tool.Name = %q, want 'read_file'", tool.Name)
	}
}

func TestWriteFileToolDefinition(t *testing.T) {
	tool := WriteFileToolDefinition()
	if tool.Name != "write_file" {
		t.Errorf("tool.Name = %q, want 'write_file'", tool.Name)
	}
}

func TestStrReplaceToolDefinition(t *testing.T) {
	tool := StrReplaceToolDefinition()
	if tool.Name != "str_replace" {
		t.Errorf("tool.Name = %q, want 'str_replace'", tool.Name)
	}
}

func TestLsToolDefinition(t *testing.T) {
	tool := LsToolDefinition()
	if tool.Name != "ls" {
		t.Errorf("tool.Name = %q, want 'ls'", tool.Name)
	}
}
