package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/user/deer-flow-go/internal/model"
)

var ErrThreadNotFound = errors.New("thread not found")

type ThreadStore struct {
	baseDir string
	mu      sync.RWMutex
}

func NewThreadStore(baseDir string) *ThreadStore {
	return &ThreadStore{
		baseDir: baseDir,
	}
}

func (s *ThreadStore) Create() (*model.Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	thread := &model.Thread{
		ThreadID:  uuid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}

	threadDir := filepath.Join(s.baseDir, "threads", thread.ThreadID)
	if err := os.MkdirAll(threadDir, 0755); err != nil {
		return nil, err
	}

	threadPath := filepath.Join(threadDir, "thread.json")
	if err := s.writeThread(threadPath, thread); err != nil {
		return nil, err
	}

	return thread, nil
}

func (s *ThreadStore) Get(threadID string) (*model.Thread, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	threadPath := filepath.Join(s.baseDir, "threads", threadID, "thread.json")

	data, err := os.ReadFile(threadPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}

	var thread model.Thread
	if err := json.Unmarshal(data, &thread); err != nil {
		return nil, err
	}

	return &thread, nil
}

func (s *ThreadStore) Delete(threadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	threadDir := filepath.Join(s.baseDir, "threads", threadID)

	if _, err := os.Stat(threadDir); os.IsNotExist(err) {
		return ErrThreadNotFound
	}

	if err := os.RemoveAll(threadDir); err != nil {
		return err
	}

	return nil
}

func (s *ThreadStore) writeThread(path string, thread *model.Thread) error {
	data, err := json.MarshalIndent(thread, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}
