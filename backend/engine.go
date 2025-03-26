package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aussiebroadwan/taboo/backend/db"
	"github.com/aussiebroadwan/taboo/backend/db/sqlc"
	"github.com/aussiebroadwan/taboo/backend/pkg/hub"
	"github.com/aussiebroadwan/taboo/backend/pkg/rng"
	"github.com/aussiebroadwan/taboo/backend/pkg/web"
	"google.golang.org/protobuf/proto"
)

type GameState int

const (
	StateDrawing GameState = iota // Round is Active
	StateWaiting                  // Round has ended
)

const (
	DrawTime = 90 * time.Second
	WaitTime = 90 * time.Second
)

type Engine struct {
	rngService *rng.RNGService
	hub        *hub.Hub

	gameId int64

	nextGameTime time.Time

	state            GameState
	currentDraw      []uint8
	currentPickIndex int
}

// NewEngine creates a new instance of the game engine as well
// as gets the initial draw.
func NewEngine(hub *hub.Hub) *Engine {
	queries := sqlc.New(db.GetDB())

	lastGameId, err := queries.GetLastGameID(context.Background())
	if err != nil {
		log.Fatalf("engine: failed to get last game_id: %v", err)
	}

	engine := &Engine{
		rngService:   rng.NewRNGService(),
		hub:          hub,
		state:        StateDrawing,
		gameId:       lastGameId + 1,
		nextGameTime: time.Now().Add(DrawTime + WaitTime),
	}

	// Capture inital Draw
	engine.currentDraw = engine.rngService.GetDraw()
	engine.currentPickIndex = 0

	if err = engine.saveGame(); err != nil {
		log.Fatalf("engine: failed to save initial game: %v", err)
	}

	return engine
}

func (e *Engine) saveGame() error {
	queries := sqlc.New(db.GetDB())

	picksStr := ""
	for idx := range e.currentDraw {
		if idx == 0 {
			picksStr = fmt.Sprintf("%d", e.currentDraw[idx])
			continue
		}

		picksStr += fmt.Sprintf(",%d", e.currentDraw[idx])
	}

	return queries.CreateGame(context.Background(), &sqlc.CreateGameParams{
		GameID: e.gameId,
		Picks:  picksStr,
	})

}

// run starts the game loop which switches between drawing and
// waiting phases.
// run is the main loop which handles state transitions and broadcasting picks.
func (e *Engine) run() {
	// Ticker for checking overall round state.
	stateTicker := time.NewTicker(1 * time.Second)
	defer stateTicker.Stop()

	var drawTicker *time.Ticker
	if e.state == StateDrawing {
		drawTicker = time.NewTicker(DrawTime / 20) // 4.5 sec intervals for picks
	}

	roundStart := time.Now()

	for {
		// Set drawChan to the ticker channel if drawTicker is active; nil otherwise.
		var drawChan <-chan time.Time
		if drawTicker != nil {
			drawChan = drawTicker.C
		} else {
			drawChan = nil
		}

		select {
		case now := <-stateTicker.C:
			switch e.state {
			case StateDrawing:
				// End drawing round if e.DrawDuration has elapsed.
				if now.Sub(roundStart) >= DrawTime {
					e.state = StateWaiting
					roundStart = now
					if drawTicker != nil {
						drawTicker.Stop()
						drawTicker = nil
					}
				}
			case StateWaiting:
				// After waiting period, start a new drawing round.
				if now.Sub(roundStart) >= WaitTime {

					// Transition to drawing phase.
					e.state = StateDrawing
					roundStart = now
					e.currentDraw = e.rngService.GetDraw()
					e.currentPickIndex = 0
					drawTicker = time.NewTicker(DrawTime / 21)

					// Update nextGameTime and increment game id as needed.
					e.nextGameTime = time.Now().Add(DrawTime + WaitTime)
					e.gameId++

					log.Printf("engine: drawing game (%d) with picks %+v", e.gameId, e.currentDraw)

					if err := e.saveGame(); err != nil {
						log.Printf("engine: failed to save game [%d] to db... continuing...: %v", e.gameId, err)
					}

					// Broadcast updated game state.
					gameStateMsg := CreateGameStateMessage(e)
					e.hub.Broadcast <- gameStateMsg
				}
			}
		case <-drawChan:
			// Only process draw ticks during drawing phase.
			if e.state == StateDrawing && e.currentPickIndex < len(e.currentDraw) {
				pick := e.currentDraw[e.currentPickIndex]
				e.currentPickIndex++
				msg := CreateNextPickMessage(pick)
				e.hub.Broadcast <- msg
			}
		}
	}
}

// CreateNextPickMessage creates a binary message for the next pick.
// Replace this with your actual protobuf encoding.
func CreateNextPickMessage(pick uint8) []byte {
	nextPick := &web.NextPick{
		PickNumber: int32(pick),
	}
	serverMsg := &web.ServerMessage{
		Payload: &web.ServerMessage_NextPick{
			NextPick: nextPick,
		},
	}
	data, err := proto.Marshal(serverMsg)
	if err != nil {
		log.Printf("engine: error marshaling NextPick: %v", err)
		return nil
	}

	return data
}

// CreateGameStateMessage creates a binary message for the current game state.
// Replace this with your actual protobuf encoding.
func CreateGameStateMessage(e *Engine) []byte {
	var state web.GameState
	switch e.state {
	case StateDrawing:
		state = web.GameState_GAME_STATE_DRAWING
	case StateWaiting:
		state = web.GameState_GAME_STATE_COMPLETED
	default:
		state = web.GameState_GAME_STATE_INVALID
	}

	var picks []int32
	if e.currentPickIndex > 0 {
		for idx := range e.currentDraw[:e.currentPickIndex] {
			picks = append(picks, int32(e.currentDraw[idx]))
		}
	}

	gameInfo := &web.GameInfo{
		GameId:       int32(e.gameId),
		State:        state,
		NextGameTime: e.nextGameTime.Format(time.RFC3339),
		Picks:        picks,
	}
	serverMsg := &web.ServerMessage{
		Payload: &web.ServerMessage_GameInfo{
			GameInfo: gameInfo,
		},
	}
	data, err := proto.Marshal(serverMsg)
	if err != nil {
		log.Printf("engine: error marshaling GameInfo: %v", err)
		return nil
	}

	return data
}

// SendCurrentGameState sends the current game state to the given client.
func (e *Engine) SendCurrentGameState(client *hub.Client) {
	msg := CreateGameStateMessage(e)
	client.Send <- msg
}
