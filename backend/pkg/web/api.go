package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/lcox74/tabo/backend/db"
	"github.com/lcox74/tabo/backend/db/sqlc"
)

// GetGameIDResponse defines the JSON structure returned for a game request.
type GetGameIDResponse struct {
	GameId int   `json:"gameId"`
	Picks  []int `json:"picks"`
}

// RegisterAPI registers the API endpoints for game-related requests.
func RegisterAPI(router *http.ServeMux) {
	router.HandleFunc("/api/game/{gameid}", func(w http.ResponseWriter, r *http.Request) {
		gameIdStr := r.PathValue("gameid")

		gameId, err := strconv.Atoi(gameIdStr)
		if err != nil {
			log.Printf("api: error converting gameid '%s' to integer: %v", gameIdStr, err)
			http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
			return
		}

		queries := sqlc.New(db.GetDB())
		row, err := queries.GetGameByGameID(context.Background(), int64(gameId))
		if err != nil || row.Picks == "" {
			log.Printf("api: error fetching game with id %d: %v", gameId, err)
			http.Error(w, "Unable to find game", http.StatusNotFound)
			return
		}

		picksStr := strings.Split(row.Picks, ",")
		resp := GetGameIDResponse{
			GameId: gameId,
			Picks:  make([]int, 0),
		}

		for _, pick := range picksStr {
			n, err := strconv.Atoi(pick)
			if err != nil {
				log.Printf("api: error converting pick '%s' to integer: %v", pick, err)
				http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
				return
			}
			resp.Picks = append(resp.Picks, n)
		}

		// Set the response header to JSON.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Encode and send the response as JSON.
		if err = json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("api: error encoding JSON response for game ID %d: %v", gameId, err)
			http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
		}
	})
}
