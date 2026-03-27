package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	return &SSEWriter{
		w:       w,
		flusher: flusher,
	}
}

func (s *SSEWriter) WriteEvent(event string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	fmt.Fprintf(s.w, "event: %s\n", event)
	fmt.Fprintf(s.w, "data: %s\n\n", dataBytes)
	s.flusher.Flush()

	return nil
}

func (s *SSEWriter) Close() {
	fmt.Fprint(s.w, "event: end\ndata: {}\n\n")
	s.flusher.Flush()
}
