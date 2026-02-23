package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEStream wraps an http.ResponseWriter for SSE communication.
type SSEStream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEStream creates a new SSE stream and sets appropriate headers.
// Returns nil if the response writer doesn't support flushing.
func NewSSEStream(w http.ResponseWriter) *SSEStream {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return &SSEStream{
		w:       w,
		flusher: flusher,
	}
}

// Send writes an SSE event with the given type and data.
func (s *SSEStream) Send(eventType string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling event data: %w", err)
	}

	// Write SSE format: event: <type>\ndata: <json>\n\n
	_, err = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventType, jsonData)
	if err != nil {
		return fmt.Errorf("writing event: %w", err)
	}

	s.flusher.Flush()
	return nil
}

// SendHeartbeat sends a heartbeat event.
func (s *SSEStream) SendHeartbeat() error {
	return s.Send("game:heartbeat", struct{}{})
}

