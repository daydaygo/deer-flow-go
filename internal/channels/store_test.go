package channels

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChannelStoreKey(t *testing.T) {
	store := NewChannelStore("")

	key := store.key("slack", "C123", "")
	if key != "slack:C123" {
		t.Errorf("expected key 'slack:C123', got %s", key)
	}

	key = store.key("slack", "C123", "T123")
	if key != "slack:C123:T123" {
		t.Errorf("expected key 'slack:C123:T123', got %s", key)
	}
}

func TestChannelStoreSetGet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "channel_store_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "store.json")
	store := NewChannelStore(storePath)

	threadID := "thread-123"
	store.SetThreadID("slack", "C123", threadID, "", "U123")

	retrieved := store.GetThreadID("slack", "C123", "")
	if retrieved != threadID {
		t.Errorf("expected threadID '%s', got '%s'", threadID, retrieved)
	}
}

func TestChannelStoreWithTopic(t *testing.T) {
	store := NewChannelStore("")

	threadID := "thread-123"
	store.SetThreadID("slack", "C123", threadID, "T456", "U123")

	retrieved := store.GetThreadID("slack", "C123", "T456")
	if retrieved != threadID {
		t.Errorf("expected threadID '%s', got '%s'", threadID, retrieved)
	}

	retrieved = store.GetThreadID("slack", "C123", "")
	if retrieved != "" {
		t.Errorf("expected empty threadID for different topic, got '%s'", retrieved)
	}
}

func TestChannelStoreRemove(t *testing.T) {
	store := NewChannelStore("")

	store.SetThreadID("slack", "C123", "thread-1", "", "U123")
	store.SetThreadID("slack", "C123", "thread-2", "T456", "U123")

	removed := store.Remove("slack", "C123", "T456")
	if !removed {
		t.Errorf("expected remove to return true")
	}

	if store.GetThreadID("slack", "C123", "T456") != "" {
		t.Errorf("expected threadID to be removed")
	}

	if store.GetThreadID("slack", "C123", "") != "thread-1" {
		t.Errorf("expected base threadID to remain")
	}
}

func TestChannelStoreRemoveAll(t *testing.T) {
	store := NewChannelStore("")

	store.SetThreadID("slack", "C123", "thread-1", "", "U123")
	store.SetThreadID("slack", "C123", "thread-2", "T456", "U123")

	removed := store.Remove("slack", "C123", "")
	if !removed {
		t.Errorf("expected remove to return true")
	}

	if store.GetThreadID("slack", "C123", "") != "" {
		t.Errorf("expected all threadIDs to be removed")
	}
	if store.GetThreadID("slack", "C123", "T456") != "" {
		t.Errorf("expected all threadIDs to be removed")
	}
}

func TestChannelStorePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "channel_store_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "store.json")

	store1 := NewChannelStore(storePath)
	threadID := "thread-123"
	store1.SetThreadID("slack", "C123", threadID, "", "U123")

	store2 := NewChannelStore(storePath)
	retrieved := store2.GetThreadID("slack", "C123", "")
	if retrieved != threadID {
		t.Errorf("expected persisted threadID '%s', got '%s'", threadID, retrieved)
	}
}

func TestChannelStoreListEntries(t *testing.T) {
	store := NewChannelStore("")
	store.SetThreadID("slack", "C123", "thread-1", "", "U123")
	store.SetThreadID("slack", "C456", "thread-2", "T789", "U456")
	store.SetThreadID("telegram", "G123", "thread-3", "", "U789")

	entries := store.ListEntries("")
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	slackEntries := store.ListEntries("slack")
	if len(slackEntries) != 2 {
		t.Errorf("expected 2 slack entries, got %d", len(slackEntries))
	}

	for _, entry := range slackEntries {
		if entry["channel_name"] != "slack" {
			t.Errorf("expected channel_name 'slack'")
		}
	}
}

func TestSplitKey(t *testing.T) {
	parts := splitKey("slack:C123")
	if len(parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(parts))
	}
	if parts[0] != "slack" || parts[1] != "C123" {
		t.Errorf("expected ['slack', 'C123'], got %v", parts)
	}

	parts = splitKey("slack:C123:T456")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts, got %d", len(parts))
	}
	if parts[0] != "slack" || parts[1] != "C123" || parts[2] != "T456" {
		t.Errorf("expected ['slack', 'C123', 'T456'], got %v", parts)
	}
}
