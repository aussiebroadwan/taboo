package sdk

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EventHandler handles SSE events from the Taboo server.
type EventHandler interface {
	OnGameState(GameStateEvent)
	OnGamePick(GamePickEvent)
	OnGameComplete(GameCompleteEvent)
	OnHeartbeat()
	OnConnect()
	OnDisconnect(error)
}

// BaseEventHandler provides default no-op implementations for EventHandler.
// Embed this in your handler to only implement the methods you need.
type BaseEventHandler struct{}

func (BaseEventHandler) OnGameState(GameStateEvent)       {}
func (BaseEventHandler) OnGamePick(GamePickEvent)         {}
func (BaseEventHandler) OnGameComplete(GameCompleteEvent) {}
func (BaseEventHandler) OnHeartbeat()                     {}
func (BaseEventHandler) OnConnect()                       {}
func (BaseEventHandler) OnDisconnect(error)               {}

// SSEClient connects to the Taboo SSE endpoint and dispatches events.
type SSEClient struct {
	baseURL        string
	handler        EventHandler
	httpClient     *http.Client
	reconnectDelay time.Duration
	maxRetries     int // 0 = unlimited
}

// SSEOption configures the SSEClient.
type SSEOption func(*SSEClient)

// WithReconnectDelay sets the delay between reconnection attempts.
func WithReconnectDelay(d time.Duration) SSEOption {
	return func(c *SSEClient) {
		c.reconnectDelay = d
	}
}

// WithMaxRetries sets the maximum number of reconnection attempts (0 = unlimited).
func WithMaxRetries(n int) SSEOption {
	return func(c *SSEClient) {
		c.maxRetries = n
	}
}

// WithSSEHTTPClient sets a custom HTTP client for the SSE connection.
func WithSSEHTTPClient(hc *http.Client) SSEOption {
	return func(c *SSEClient) {
		c.httpClient = hc
	}
}

// NewSSEClient creates a new SSE client.
func NewSSEClient(baseURL string, handler EventHandler, opts ...SSEOption) *SSEClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	c := &SSEClient{
		baseURL:        baseURL,
		handler:        handler,
		httpClient:     &http.Client{},
		reconnectDelay: 5 * time.Second,
		maxRetries:     0,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Connect establishes an SSE connection and processes events.
// It blocks until the context is cancelled, automatically reconnecting on errors.
func (c *SSEClient) Connect(ctx context.Context) error {
	retries := 0
	for {
		err := c.connect(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		c.handler.OnDisconnect(err)
		retries++

		if c.maxRetries > 0 && retries >= c.maxRetries {
			return fmt.Errorf("max retries (%d) exceeded: %w", c.maxRetries, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.reconnectDelay):
			// Continue to reconnect
		}
	}
}

func (c *SSEClient) connect(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/events", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	c.handler.OnConnect()

	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	var data strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if eventType != "" && data.Len() > 0 {
				c.dispatchEvent(eventType, data.String())
			}
			eventType = ""
			data.Reset()
			continue
		}

		if after, ok := strings.CutPrefix(line, "event:"); ok {
			eventType = strings.TrimSpace(after)
		} else if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteString("\n")
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
		// Ignore other fields (id, retry, comments)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}

	return nil
}

func (c *SSEClient) dispatchEvent(eventType, data string) {
	switch eventType {
	case EventGameState:
		var e GameStateEvent
		if json.Unmarshal([]byte(data), &e) == nil {
			c.handler.OnGameState(e)
		}
	case EventGamePick:
		var e GamePickEvent
		if json.Unmarshal([]byte(data), &e) == nil {
			c.handler.OnGamePick(e)
		}
	case EventGameComplete:
		var e GameCompleteEvent
		if json.Unmarshal([]byte(data), &e) == nil {
			c.handler.OnGameComplete(e)
		}
	case EventGameHeartbeat:
		c.handler.OnHeartbeat()
	}
}
