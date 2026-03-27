package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type UserMemory struct {
	WorkContext        string `json:"workContext,omitempty"`
	PersonalContext    string `json:"personalContext,omitempty"`
	TopOfMind          string `json:"topOfMind,omitempty"`
	RecentMonths       string `json:"recentMonths,omitempty"`
	EarlierContext     string `json:"earlierContext,omitempty"`
	LongTermBackground string `json:"longTermBackground,omitempty"`
	Facts              []Fact `json:"facts,omitempty"`
}

type Fact struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Category   string  `json:"category,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	CreatedAt  string  `json:"createdAt,omitempty"`
	Source     string  `json:"source,omitempty"`
}

type MemoryStore struct {
	storagePath string
	mu          sync.RWMutex
	cache       *UserMemory
}

func NewMemoryStore(storagePath string) *MemoryStore {
	return &MemoryStore{
		storagePath: storagePath,
	}
}

func (s *MemoryStore) Load() (*UserMemory, error) {
	s.mu.RLock()
	if s.cache != nil {
		defer s.mu.RUnlock()
		return s.cache, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil {
		return s.cache, nil
	}

	data, err := os.ReadFile(s.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			mem := &UserMemory{}
			s.cache = mem
			return mem, nil
		}
		return nil, err
	}

	var mem UserMemory
	if err := json.Unmarshal(data, &mem); err != nil {
		return nil, err
	}

	s.cache = &mem
	return &mem, nil
}

func (s *MemoryStore) Save(mem *UserMemory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.storagePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(mem, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := s.storagePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, s.storagePath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	s.cache = mem
	return nil
}

func (s *MemoryStore) Reload() (*UserMemory, error) {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()

	return s.Load()
}
