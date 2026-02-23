package http

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/service"
	"github.com/aussiebroadwan/taboo/sdk"
)

// readSSEEvent reads a single SSE event from the reader.
// Returns event type, data, and any error.
func readSSEEvent(reader *bufio.Reader) (eventType, data string, err error) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", "", err
		}

		line = strings.TrimSuffix(line, "\n")

		if line == "" {
			// Empty line marks end of event
			if eventType != "" || data != "" {
				return eventType, data, nil
			}
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			data = strings.TrimPrefix(line, "data: ")
		}
	}
}

func TestSSE_ConnectionHeaders(t *testing.T) {
	store := newMockStore()
	cfg := config.Default()
	// Use a very short heartbeat for testing
	cfg.Server.SSEHeartbeat = config.Duration(50 * time.Millisecond)
	gameService := service.NewGameService(store, &cfg.Game)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(cfg, logger, store, gameService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)

	// Use custom writer that supports SetWriteDeadline
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	w := newSSEResponseWriter(pw)

	// Run handler in goroutine since it blocks
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	go server.handleEvents(w, req)

	// Wait for headers to be set (WriteHeader called)
	w.WaitForHeaders()

	// Check headers - safe to read now since WriteHeader has been called
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type 'text/event-stream', got %q", ct)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control 'no-cache', got %q", cc)
	}
	if conn := w.Header().Get("Connection"); conn != "keep-alive" {
		t.Errorf("expected Connection 'keep-alive', got %q", conn)
	}
}

func TestSSE_ReceiveEvent(t *testing.T) {
	store := newMockStore()
	cfg := config.Default()
	cfg.Server.SSEHeartbeat = config.Duration(10 * time.Second) // Long heartbeat to avoid interference
	gameService := service.NewGameService(store, &cfg.Game)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(cfg, logger, store, gameService, nil)

	// Use a pipe to read SSE events
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	w := newSSEResponseWriter(pw)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil).WithContext(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleEvents(w, req)
	}()

	// Wait for headers to be set up before broadcasting
	w.WaitForHeaders()
	// Small delay to ensure subscription is established after headers
	time.Sleep(10 * time.Millisecond)

	// Broadcast an event
	gameService.BroadcastPick(42)

	// Read the event
	reader := bufio.NewReader(pr)
	eventType, data, err := readSSEEvent(reader)
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}

	if eventType != sdk.EventGamePick {
		t.Errorf("expected event type %q, got %q", sdk.EventGamePick, eventType)
	}
	if !strings.Contains(data, "42") {
		t.Errorf("expected data to contain '42', got %q", data)
	}

	cancel()
	wg.Wait()
}

func TestSSE_MultipleEvents(t *testing.T) {
	store := newMockStore()
	cfg := config.Default()
	cfg.Server.SSEHeartbeat = config.Duration(10 * time.Second)
	gameService := service.NewGameService(store, &cfg.Game)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(cfg, logger, store, gameService, nil)

	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	w := newSSEResponseWriter(pw)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil).WithContext(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleEvents(w, req)
	}()

	w.WaitForHeaders()
	// Small delay to ensure subscription is established after headers
	time.Sleep(10 * time.Millisecond)

	// Broadcast multiple events
	gameService.BroadcastPick(1)
	gameService.BroadcastPick(2)
	gameService.BroadcastPick(3)

	reader := bufio.NewReader(pr)

	// Read all three events
	picks := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		eventType, data, err := readSSEEvent(reader)
		if err != nil {
			t.Fatalf("failed to read event %d: %v", i, err)
		}
		if eventType != sdk.EventGamePick {
			t.Errorf("event %d: expected type %q, got %q", i, sdk.EventGamePick, eventType)
		}
		picks = append(picks, data)
	}

	// Verify order
	if !strings.Contains(picks[0], "1") {
		t.Errorf("first event should contain '1', got %q", picks[0])
	}
	if !strings.Contains(picks[1], "2") {
		t.Errorf("second event should contain '2', got %q", picks[1])
	}
	if !strings.Contains(picks[2], "3") {
		t.Errorf("third event should contain '3', got %q", picks[2])
	}

	cancel()
	wg.Wait()
}

func TestSSE_Heartbeat(t *testing.T) {
	store := newMockStore()
	cfg := config.Default()
	cfg.Server.SSEHeartbeat = config.Duration(50 * time.Millisecond) // Very short for testing
	gameService := service.NewGameService(store, &cfg.Game)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(cfg, logger, store, gameService, nil)

	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	w := newSSEResponseWriter(pw)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil).WithContext(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleEvents(w, req)
	}()

	reader := bufio.NewReader(pr)

	// Wait for heartbeat
	done := make(chan struct{})
	go func() {
		eventType, _, err := readSSEEvent(reader)
		if err != nil {
			t.Errorf("failed to read heartbeat: %v", err)
		}
		if eventType != "game:heartbeat" {
			t.Errorf("expected heartbeat event, got %q", eventType)
		}
		close(done)
	}()

	select {
	case <-done:
		// Got heartbeat
	case <-time.After(200 * time.Millisecond):
		t.Error("timeout waiting for heartbeat")
	}

	cancel()
	wg.Wait()
}

func TestSSE_ClientDisconnect(t *testing.T) {
	store := newMockStore()
	cfg := config.Default()
	cfg.Server.SSEHeartbeat = config.Duration(10 * time.Second)
	gameService := service.NewGameService(store, &cfg.Game)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(cfg, logger, store, gameService, nil)

	pr, pw := io.Pipe()

	w := newSSEResponseWriter(pw)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil).WithContext(ctx)

	handlerDone := make(chan struct{})
	go func() {
		server.handleEvents(w, req)
		close(handlerDone)
	}()

	w.WaitForHeaders()

	// Disconnect client
	cancel()
	pr.Close()
	pw.Close()

	// Handler should exit
	select {
	case <-handlerDone:
		// Handler exited as expected
	case <-time.After(500 * time.Millisecond):
		t.Error("handler did not exit after client disconnect")
	}
}

func TestSSE_MultipleClients(t *testing.T) {
	store := newMockStore()
	cfg := config.Default()
	cfg.Server.SSEHeartbeat = config.Duration(10 * time.Second)
	gameService := service.NewGameService(store, &cfg.Game)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(cfg, logger, store, gameService, nil)

	const clientCount = 3
	readers := make([]*bufio.Reader, clientCount)
	cancels := make([]context.CancelFunc, clientCount)
	writers := make([]*sseResponseWriter, clientCount)
	var wg sync.WaitGroup

	for i := 0; i < clientCount; i++ {
		pr, pw := io.Pipe()
		defer pr.Close()
		defer pw.Close()

		w := newSSEResponseWriter(pw)
		writers[i] = w

		ctx, cancel := context.WithCancel(context.Background())
		cancels[i] = cancel
		readers[i] = bufio.NewReader(pr)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil).WithContext(ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()
			server.handleEvents(w, req)
		}()
	}

	// Wait for all clients to be ready
	for _, w := range writers {
		w.WaitForHeaders()
	}
	// Small delay to ensure subscriptions are established after headers
	time.Sleep(10 * time.Millisecond)

	// Broadcast event
	gameService.BroadcastComplete(123)

	// All clients should receive it
	for i, reader := range readers {
		eventType, data, err := readSSEEvent(reader)
		if err != nil {
			t.Errorf("client %d: failed to read event: %v", i, err)
			continue
		}
		if eventType != sdk.EventGameComplete {
			t.Errorf("client %d: expected type %q, got %q", i, sdk.EventGameComplete, eventType)
		}
		if !strings.Contains(data, "123") {
			t.Errorf("client %d: expected data to contain '123', got %q", i, data)
		}
	}

	// Cleanup
	for _, cancel := range cancels {
		cancel()
	}
	wg.Wait()
}

// sseResponseWriter implements http.ResponseWriter for SSE testing.
type sseResponseWriter struct {
	header      http.Header
	writer      io.Writer
	status      int
	headersDone chan struct{} // closed when WriteHeader is called
	headersOnce sync.Once
}

func newSSEResponseWriter(w io.Writer) *sseResponseWriter {
	return &sseResponseWriter{
		header:      make(http.Header),
		writer:      w,
		headersDone: make(chan struct{}),
	}
}

func (w *sseResponseWriter) Header() http.Header {
	return w.header
}

func (w *sseResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *sseResponseWriter) WriteHeader(status int) {
	w.status = status
	w.headersOnce.Do(func() {
		close(w.headersDone)
	})
}

func (w *sseResponseWriter) Flush() {
	// No-op for pipe writer
}

// SetWriteDeadline implements the interface required by http.ResponseController.
func (w *sseResponseWriter) SetWriteDeadline(deadline time.Time) error {
	return nil // No-op for testing
}

// SetReadDeadline implements the interface required by http.ResponseController.
func (w *sseResponseWriter) SetReadDeadline(deadline time.Time) error {
	return nil // No-op for testing
}

// WaitForHeaders blocks until WriteHeader has been called.
func (w *sseResponseWriter) WaitForHeaders() {
	<-w.headersDone
}
