package channels

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ThreadMapping struct {
	ThreadID  string    `json:"thread_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChannelStore struct {
	path string
	data map[string]ThreadMapping
	mu   sync.RWMutex
}

func NewChannelStore(path string) *ChannelStore {
	s := &ChannelStore{
		path: path,
		data: make(map[string]ThreadMapping),
	}
	s.load()
	return s
}

func (s *ChannelStore) load() {
	if s.path == "" {
		return
	}
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &s.data); err != nil {
		return
	}
}

func (s *ChannelStore) save() {
	if s.path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return
	}
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return
	}
	os.Rename(tmpPath, s.path)
}

func (s *ChannelStore) key(channelName, chatID, topicID string) string {
	if topicID != "" {
		return channelName + ":" + chatID + ":" + topicID
	}
	return channelName + ":" + chatID
}

func (s *ChannelStore) GetThreadID(channelName, chatID, topicID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := s.key(channelName, chatID, topicID)
	if mapping, ok := s.data[key]; ok {
		return mapping.ThreadID
	}
	return ""
}

func (s *ChannelStore) SetThreadID(channelName, chatID, threadID, topicID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.key(channelName, chatID, topicID)
	now := time.Now()
	var createdAt time.Time
	if existing, exists := s.data[key]; exists {
		createdAt = existing.CreatedAt
	} else {
		createdAt = now
	}
	s.data[key] = ThreadMapping{
		ThreadID:  threadID,
		UserID:    userID,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}
	s.save()
}

func (s *ChannelStore) Remove(channelName, chatID, topicID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if topicID != "" {
		key := s.key(channelName, chatID, topicID)
		if _, ok := s.data[key]; ok {
			delete(s.data, key)
			s.save()
			return true
		}
		return false
	}
	prefix := s.key(channelName, chatID, "") + ":"
	deleted := false
	for key := range s.data {
		if key == s.key(channelName, chatID, "") || (len(key) > len(prefix) && key[:len(prefix)] == prefix) {
			delete(s.data, key)
			deleted = true
		}
	}
	if deleted {
		s.save()
	}
	return deleted
}

func (s *ChannelStore) ListEntries(channelName string) []map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]map[string]any, 0)
	for key, mapping := range s.data {
		parts := splitKey(key)
		if channelName != "" && parts[0] != channelName {
			continue
		}
		entry := map[string]any{
			"channel_name": parts[0],
			"chat_id":      parts[1],
			"thread_id":    mapping.ThreadID,
			"user_id":      mapping.UserID,
			"created_at":   mapping.CreatedAt,
			"updated_at":   mapping.UpdatedAt,
		}
		if len(parts) > 2 {
			entry["topic_id"] = parts[2]
		}
		result = append(result, entry)
	}
	return result
}

func splitKey(key string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			parts = append(parts, key[start:i])
			start = i + 1
			if len(parts) == 2 {
				parts = append(parts, key[start:])
				return parts
			}
		}
	}
	if start < len(key) {
		parts = append(parts, key[start:])
	}
	return parts
}
