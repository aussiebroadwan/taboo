package sdk

import "time"

// SSE event type constants.
const (
	EventGameState     = "game:state"
	EventGamePick      = "game:pick"
	EventGameComplete  = "game:complete"
	EventGameHeartbeat = "game:heartbeat"
)

// GameStateEvent is sent when a new game starts or client connects.
type GameStateEvent struct {
	GameID   int64     `json:"game_id"`
	Picks    Picks     `json:"picks"`
	NextGame time.Time `json:"next_game"`
}

// GamePickEvent is sent when a new number is picked.
type GamePickEvent struct {
	Pick uint8 `json:"pick"`
}

// GameCompleteEvent is sent when a game finishes.
type GameCompleteEvent struct {
	GameID int64 `json:"game_id"`
}

// HeartbeatEvent is sent periodically to keep the connection alive.
type HeartbeatEvent struct{}
