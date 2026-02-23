package sdk_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aussiebroadwan/taboo/sdk"
)

func TestClient_ListGames(t *testing.T) {
	games := []sdk.Game{
		{ID: 1, Picks: sdk.Picks{1, 2, 3}, CreatedAt: time.Now()},
		{ID: 2, Picks: sdk.Picks{4, 5, 6}, CreatedAt: time.Now()},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/games" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		resp := sdk.GameListResponse{Games: games}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	resp, err := client.ListGames(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Games) != 2 {
		t.Errorf("expected 2 games, got %d", len(resp.Games))
	}
	if resp.Games[0].ID != 1 {
		t.Errorf("expected game ID 1, got %d", resp.Games[0].ID)
	}
}

func TestClient_ListGames_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		limit := r.URL.Query().Get("limit")

		if cursor != "100" {
			t.Errorf("expected cursor=100, got %s", cursor)
		}
		if limit != "50" {
			t.Errorf("expected limit=50, got %s", limit)
		}

		resp := sdk.GameListResponse{Games: []sdk.Game{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	_, err := client.ListGames(context.Background(), &sdk.ListGamesOptions{
		Cursor: new(int64(100)),
		Limit:  new(50),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetGame(t *testing.T) {
	game := sdk.Game{ID: 42, Picks: sdk.Picks{1, 2, 3}, CreatedAt: time.Now()}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/games/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	result, err := client.GetGame(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 42 {
		t.Errorf("expected game ID 42, got %d", result.ID)
	}
}

func TestClient_GetGame_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(sdk.ErrorResponse{
			Error: sdk.ErrorDetail{
				Code:    "not_found",
				Message: "game 999 not found",
			},
		})
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	_, err := client.GetGame(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *sdk.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("expected code 'not_found', got %s", apiErr.Code)
	}
}

func TestClient_WithTimeout(t *testing.T) {
	client := sdk.NewClient("http://localhost", sdk.WithTimeout(5*time.Second))
	if client == nil {
		t.Fatal("expected client, got nil")
	}
}
