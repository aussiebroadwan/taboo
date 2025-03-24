
-- name: CreateGame :exec
INSERT INTO games (game_id, picks)
VALUES (?, ?);

-- name: GetGameByGameID :one
SELECT game_id, picks, created_at
FROM games
WHERE game_id = ?;

-- name: GetLastGameID :one
SELECT COALESCE(MAX(game_id), 0) + 0 AS last_game_id
FROM games;
