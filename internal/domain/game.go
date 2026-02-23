package domain

import "time"

// Game represents a single game round with its picks.
type Game struct {
	ID        int64     `json:"id"`
	Picks     []uint8   `json:"picks"`
	CreatedAt time.Time `json:"created_at"`
}

// NewGame creates a new Game with the given ID and picks.
func NewGame(id int64, picks []uint8) *Game {
	return &Game{
		ID:        id,
		Picks:     picks,
		CreatedAt: time.Now(),
	}
}
