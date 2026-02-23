package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/aussiebroadwan/taboo/sdk"
)

type testHandler struct {
	sdk.BaseEventHandler
	mu          sync.Mutex
	states      []sdk.GameStateEvent
	picks       []sdk.GamePickEvent
	completes   []sdk.GameCompleteEvent
	heartbeats  int
	connects    int
	disconnects int
}

func (h *testHandler) OnGameState(e sdk.GameStateEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.states = append(h.states, e)
}

func (h *testHandler) OnGamePick(e sdk.GamePickEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.picks = append(h.picks, e)
}

func (h *testHandler) OnGameComplete(e sdk.GameCompleteEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.completes = append(h.completes, e)
}

func (h *testHandler) OnHeartbeat() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.heartbeats++
}

func (h *testHandler) OnConnect() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connects++
}

func (h *testHandler) OnDisconnect(error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.disconnects++
}

func TestSSEClient_Connect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("expected Flusher")
			return
		}

		// Send a game state event
		fmt.Fprintf(w, "event: game:state\n")
		fmt.Fprintf(w, "data: {\"game_id\":1,\"picks\":[1,2,3],\"next_game\":\"2024-01-01T00:00:00Z\"}\n\n")
		flusher.Flush()

		// Send a pick event
		fmt.Fprintf(w, "event: game:pick\n")
		fmt.Fprintf(w, "data: {\"pick\":42}\n\n")
		flusher.Flush()

		// Send a complete event
		fmt.Fprintf(w, "event: game:complete\n")
		fmt.Fprintf(w, "data: {\"game_id\":1}\n\n")
		flusher.Flush()

		// Send a heartbeat
		fmt.Fprintf(w, "event: game:heartbeat\n")
		fmt.Fprintf(w, "data: {}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	handler := &testHandler{}
	client := sdk.NewSSEClient(server.URL, handler, sdk.WithMaxRetries(1))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Connect will return when context is cancelled or max retries exceeded
	_ = client.Connect(ctx)

	handler.mu.Lock()
	defer handler.mu.Unlock()

	if handler.connects != 1 {
		t.Errorf("expected 1 connect, got %d", handler.connects)
	}
	if len(handler.states) != 1 {
		t.Errorf("expected 1 state event, got %d", len(handler.states))
	}
	if len(handler.picks) != 1 {
		t.Errorf("expected 1 pick event, got %d", len(handler.picks))
	}
	if handler.picks[0].Pick != 42 {
		t.Errorf("expected pick 42, got %d", handler.picks[0].Pick)
	}
	if len(handler.completes) != 1 {
		t.Errorf("expected 1 complete event, got %d", len(handler.completes))
	}
	if handler.heartbeats != 1 {
		t.Errorf("expected 1 heartbeat, got %d", handler.heartbeats)
	}
}

func TestChannelHandler(t *testing.T) {
	handler := sdk.NewChannelHandler(10)

	// Simulate events
	handler.OnConnect()
	handler.OnGamePick(sdk.GamePickEvent{Pick: 7})
	handler.OnGameState(sdk.GameStateEvent{GameID: 1})
	handler.OnHeartbeat()

	// Check connected signal
	select {
	case <-handler.Connected():
		// OK
	default:
		t.Error("expected connect signal")
	}

	// Check events
	events := []any{}
	for i := 0; i < 3; i++ {
		select {
		case e := <-handler.Events():
			events = append(events, e)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	// Verify event types
	if _, ok := events[0].(sdk.GamePickEvent); !ok {
		t.Errorf("expected GamePickEvent, got %T", events[0])
	}
	if _, ok := events[1].(sdk.GameStateEvent); !ok {
		t.Errorf("expected GameStateEvent, got %T", events[1])
	}
	if _, ok := events[2].(sdk.HeartbeatEvent); !ok {
		t.Errorf("expected HeartbeatEvent, got %T", events[2])
	}
}

func TestBaseEventHandler(t *testing.T) {
	// Just verify BaseEventHandler methods don't panic
	h := sdk.BaseEventHandler{}
	h.OnGameState(sdk.GameStateEvent{})
	h.OnGamePick(sdk.GamePickEvent{})
	h.OnGameComplete(sdk.GameCompleteEvent{})
	h.OnHeartbeat()
	h.OnConnect()
	h.OnDisconnect(nil)
}
