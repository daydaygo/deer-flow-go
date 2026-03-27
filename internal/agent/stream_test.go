package agent

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSSEWriter(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	if sse == nil {
		t.Fatal("expected SSEWriter to be created")
	}

	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type 'text/event-stream', got %s", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("expected Cache-Control 'no-cache', got %s", w.Header().Get("Cache-Control"))
	}
}

func TestSSEWriter_WriteEvent(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	data := map[string]string{"message": "hello"}
	err := sse.WriteEvent("values", data)
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "event: values\n") {
		t.Error("expected event line in output")
	}
	if !strings.Contains(body, `"message":"hello"`) {
		t.Error("expected data in output")
	}
}

func TestSSEWriter_WriteEvent_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	err := sse.WriteEvent("end", nil)
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "event: end\n") {
		t.Error("expected event line in output")
	}
	if !strings.Contains(body, "data: null") {
		t.Error("expected null data in output")
	}
}

func TestSSEWriter_Close(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.Close()

	body := w.Body.String()
	if !strings.Contains(body, "event: end\n") {
		t.Error("expected end event in output")
	}
	if !strings.Contains(body, "data: {}\n") {
		t.Error("expected empty object data in output")
	}
}

func TestSSEWriter_MultipleEvents(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	err := sse.WriteEvent("values", map[string]any{"messages": []string{"Hello"}})
	if err != nil {
		t.Fatalf("failed to write first event: %v", err)
	}

	err = sse.WriteEvent("messages-tuple", map[string]any{"message": map[string]string{"role": "assistant", "content": "Hi"}})
	if err != nil {
		t.Fatalf("failed to write second event: %v", err)
	}

	sse.Close()

	body := w.Body.String()
	events := strings.Count(body, "event:")
	if events != 3 {
		t.Errorf("expected 3 events, got %d", events)
	}

	if !strings.Contains(body, `"messages":["Hello"]`) {
		t.Error("expected messages in first event")
	}
}

func TestSSEWriter_EventFormat(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	data := map[string]string{"key": "value"}
	err := sse.WriteEvent("test", data)
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	body := w.Body.String()
	expectedFormat := "event: test\ndata: {\"key\":\"value\"}\n\n"
	if !bytes.Contains([]byte(body), []byte(expectedFormat)) {
		t.Errorf("expected SSE format:\n%s\ngot:\n%s", expectedFormat, body)
	}
}
