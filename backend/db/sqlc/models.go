// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package sqlc

import (
	"time"
)

type Game struct {
	ID        int64      `json:"id"`
	GameID    int64      `json:"game_id"`
	CreatedAt *time.Time `json:"created_at"`
	Picks     string     `json:"picks"`
}
