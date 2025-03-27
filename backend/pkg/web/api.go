package web

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aussiebroadwan/taboo/backend/db"
	"github.com/aussiebroadwan/taboo/backend/db/sqlc"
)

// GameIDResponse defines the JSON structure returned for a game request.
type GameIDResponse struct {
	GameId    int        `json:"gameId" example:"100"`
	CreatedAt *time.Time `json:"createdAt" example:"2025-01-01T12:00:00Z"`
	Picks     []int      `json:"picks" example:"1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"`
}

// RegisterAPI registers the API endpoints for game-related requests.
func RegisterAPI(router *http.ServeMux) {
	router.HandleFunc("GET /api/game/latest", GetLatestGameResults)
	router.HandleFunc("GET /api/game/range", GetGameRangeResults)
	router.HandleFunc("GET /api/game/{gameid}", GetSpecificGameResults)

	router.HandleFunc("POST /api/token", PostDiscordAuthToken)

	router.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "API endpoint not found", http.StatusNotFound)
	})

}

// GetLatestGameResults handles GET requests to fetch the latest game's results.
//
// @Summary      Get latest game results
// @Description  Retrieves the most recent game ID, its creation time, and the list of picks.
// @Tags         game
// @Produce      json
// @Success      200  {object}  GameIDResponse  "Latest game result"
// @Failure      404  {string}  string          "Unable to find game"
// @Failure      500  {string}  string          "Unexpected error responding"
// @Router       /api/game/latest [get]
func GetLatestGameResults(w http.ResponseWriter, r *http.Request) {
	queries := sqlc.New(db.GetDB())

	row, err := queries.GetGameByLastGameID(context.Background())
	if err != nil || row.Picks == "" {
		log.Printf("api: error fetching last game: %v", err)
		http.Error(w, "Unable to find game", http.StatusNotFound)
		return
	}

	picksStr := strings.Split(row.Picks, ",")
	resp := GameIDResponse{
		GameId:    int(row.GameID),
		CreatedAt: row.CreatedAt,
		Picks:     make([]int, 0),
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
		log.Printf("api: error encoding JSON response for last game: %v", err)
		http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
	}

}

// GetSpecificGameResults handles GET requests to fetch a specific game's results by its ID.
//
// @Summary      Get specific game results
// @Description  Retrieves the results of a specific game given its ID, including its creation time and picks.
// @Tags         game
// @Produce      json
// @Param        gameid  path      int             true  "Game ID"
// @Success      200     {object}  GameIDResponse  "Game result by ID"
// @Failure      404     {string}  string          "Unable to find game"
// @Failure      500     {string}  string          "Unexpected error responding"
// @Router       /api/game/{gameid} [get]
func GetSpecificGameResults(w http.ResponseWriter, r *http.Request) {
	gameIdStr := r.PathValue("gameid")

	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		log.Printf("api: error converting gameid '%s' to integer: %v", gameIdStr, err)
		http.Error(w, "API endpoint not found", http.StatusNotFound)
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
	resp := GameIDResponse{
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
}

// GetGameRangeResults handles GET requests to fetch a range of games.
//
// @Summary      Get a range of game results
// @Description  Retrieves a list of game results starting from a given game ID.
// @Tags         game
// @Produce      json
// @Param        start  query     int  true   "Starting game ID"
// @Param        count  query     int  true   "Number of games to return (max 100)"
// @Success      200    {array}   GameIDResponse  "List of game results"
// @Failure      400    {string}  string           "Invalid query parameters"
// @Failure      500    {string}  string           "Unexpected error responding"
// @Router       /api/game/range [get]
func GetGameRangeResults(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	countStr := r.URL.Query().Get("count")

	start, err := strconv.Atoi(startStr)
	if err != nil || start < 1 {
		http.Error(w, "Invalid start parameter", http.StatusBadRequest)
		return
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 100 {
		http.Error(w, "Count must be between 1 and 100", http.StatusBadRequest)
		return
	}

	queries := sqlc.New(db.GetDB())
	rows, err := queries.GetGamesByRange(context.Background(), &sqlc.GetGamesByRangeParams{
		Start: int64(start),
		Limit: int64(count),
	})
	if err != nil {
		log.Printf("api: error fetching games from %d: %v", start, err)
		http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
		return
	}

	var results []GameIDResponse

	for _, row := range rows {
		picksStr := strings.Split(row.Picks, ",")
		picks := make([]int, 0, len(picksStr))

		for _, pick := range picksStr {
			n, err := strconv.Atoi(pick)
			if err != nil {
				log.Printf("api: error converting pick '%s' to integer: %v", pick, err)
				http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
				return
			}
			picks = append(picks, n)
		}

		results = append(results, GameIDResponse{
			GameId:    int(row.GameID),
			CreatedAt: row.CreatedAt,
			Picks:     picks,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("api: error encoding JSON response: %v", err)
		http.Error(w, "Unexpected error responding", http.StatusInternalServerError)
	}
}

// PostDiscordAuthToken handles POST requests to exchange a Discord code for
// an access token. This is needed for the frontend application to use the
// DiscordSDK properly.
func PostDiscordAuthToken(w http.ResponseWriter, r *http.Request) {
	// Define request and response structures
	type requestBody struct {
		Code string `json:"code"`
	}
	type responseToken struct {
		AccessToken string `json:"access_token"`
	}

	// Parse incoming JSON
	var req requestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Code) == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Construct form values for Discord token exchange
	form := url.Values{}
	form.Set("client_id", os.Getenv("DISCORD_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("DISCORD_CLIENT_SECRET"))
	form.Set("grant_type", "authorization_code")
	form.Set("code", req.Code)

	resp, err := http.PostForm("https://discord.com/api/oauth2/token", form)
	if err != nil {
		log.Printf("api: error contacting Discord: %v", err)
		http.Error(w, "Failed to contact Discord", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("api: Discord token exchange failed [%d]: %s", resp.StatusCode, string(body))
		http.Error(w, "Discord API error", http.StatusBadGateway)
		return
	}

	var token responseToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log.Printf("api: error decoding Discord token response: %v", err)
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	// Send token response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		log.Printf("api: error writing token JSON response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}
