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

var ErrRunNotFound = errors.New("run not found")

type RunStore struct {
	baseDir string
	mu      sync.RWMutex
}

func NewRunStore(baseDir string) *RunStore {
	return &RunStore{
		baseDir: baseDir,
	}
}

func (s *RunStore) Create(threadID, assistantID string, input map[string]any) (*model.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	run := &model.Run{
		RunID:       uuid.New().String(),
		ThreadID:    threadID,
		AssistantID: assistantID,
		Status:      model.RunStatusPending,
		Input:       input,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	threadDir := filepath.Join(s.baseDir, "threads", threadID)
	runsDir := filepath.Join(threadDir, "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		return nil, err
	}

	runPath := filepath.Join(runsDir, run.RunID+".json")
	if err := s.writeRun(runPath, run); err != nil {
		return nil, err
	}

	return run, nil
}

func (s *RunStore) Get(threadID, runID string) (*model.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runPath := filepath.Join(s.baseDir, "threads", threadID, "runs", runID+".json")

	data, err := os.ReadFile(runPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}

	var run model.Run
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, err
	}

	return &run, nil
}

func (s *RunStore) Update(threadID, runID string, updates func(*model.Run)) (*model.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	runPath := filepath.Join(s.baseDir, "threads", threadID, "runs", runID+".json")

	data, err := os.ReadFile(runPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}

	var run model.Run
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, err
	}

	updates(&run)
	run.UpdatedAt = time.Now()

	if err := s.writeRun(runPath, &run); err != nil {
		return nil, err
	}

	return &run, nil
}

func (s *RunStore) List(threadID string) ([]*model.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runsDir := filepath.Join(s.baseDir, "threads", threadID, "runs")

	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.Run{}, nil
		}
		return nil, err
	}

	var runs []*model.Run
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		runPath := filepath.Join(runsDir, entry.Name())
		data, err := os.ReadFile(runPath)
		if err != nil {
			continue
		}

		var run model.Run
		if err := json.Unmarshal(data, &run); err != nil {
			continue
		}

		runs = append(runs, &run)
	}

	return runs, nil
}

func (s *RunStore) writeRun(path string, run *model.Run) error {
	data, err := json.MarshalIndent(run, "", "  ")
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
