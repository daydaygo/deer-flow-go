package agent

import (
	"sync"
	"time"

	"github.com/user/deer-flow-go/internal/model"
)

type StateManager struct {
	states map[string]*model.ThreadState
	mu     sync.RWMutex
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]*model.ThreadState),
	}
}

func (s *StateManager) Get(threadID string) *model.ThreadState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[threadID]
}

func (s *StateManager) GetOrCreate(threadID string) *model.ThreadState {
	s.mu.RLock()
	if state, ok := s.states[threadID]; ok {
		s.mu.RUnlock()
		return state
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if state, ok := s.states[threadID]; ok {
		return state
	}

	now := time.Now()
	state := &model.ThreadState{
		ThreadID:  threadID,
		Messages:  []model.Message{},
		Artifacts: []model.Artifact{},
		Context:   make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.states[threadID] = state
	return state
}

func (s *StateManager) Update(threadID string, state *model.ThreadState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[threadID] = state
}
