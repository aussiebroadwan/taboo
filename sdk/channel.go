package sdk

// ChannelHandler implements EventHandler by sending events to a channel.
// This allows for a select-based event loop instead of callbacks.
type ChannelHandler struct {
	events      chan any
	connected   chan struct{}
	disconnects chan error
}

// NewChannelHandler creates a new channel-based event handler.
// The buffer parameter sets the channel buffer size.
func NewChannelHandler(buffer int) *ChannelHandler {
	return &ChannelHandler{
		events:      make(chan any, buffer),
		connected:   make(chan struct{}, 1),
		disconnects: make(chan error, 1),
	}
}

// Events returns a channel that receives all game events.
// Events are one of: GameStateEvent, GamePickEvent, GameCompleteEvent, HeartbeatEvent.
func (h *ChannelHandler) Events() <-chan any {
	return h.events
}

// Connected returns a channel that signals when the connection is established.
func (h *ChannelHandler) Connected() <-chan struct{} {
	return h.connected
}

// Disconnects returns a channel that receives disconnect errors.
func (h *ChannelHandler) Disconnects() <-chan error {
	return h.disconnects
}

// Close closes all channels. Call this when done with the handler.
func (h *ChannelHandler) Close() {
	close(h.events)
	close(h.connected)
	close(h.disconnects)
}

// EventHandler interface implementation

func (h *ChannelHandler) OnGameState(e GameStateEvent) {
	select {
	case h.events <- e:
	default:
	}
}

func (h *ChannelHandler) OnGamePick(e GamePickEvent) {
	select {
	case h.events <- e:
	default:
	}
}

func (h *ChannelHandler) OnGameComplete(e GameCompleteEvent) {
	select {
	case h.events <- e:
	default:
	}
}

func (h *ChannelHandler) OnHeartbeat() {
	select {
	case h.events <- HeartbeatEvent{}:
	default:
	}
}

func (h *ChannelHandler) OnConnect() {
	select {
	case h.connected <- struct{}{}:
	default:
	}
}

func (h *ChannelHandler) OnDisconnect(err error) {
	select {
	case h.disconnects <- err:
	default:
	}
}
