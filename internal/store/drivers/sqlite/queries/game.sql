
-- name: CreateGame :exec
INSERT INTO games (game_id, picks)
VALUES (?, ?);

-- name: GetGameByGameID :one
SELECT game_id, picks, created_at
FROM games
WHERE game_id = ?;

-- name: GetLatestGame :one
SELECT game_id, picks, created_at
FROM games
ORDER BY game_id DESC
LIMIT 1;

-- name: GetGamesByRange :many
SELECT game_id, picks, created_at
FROM games
WHERE game_id >= sqlc.arg('start')
ORDER BY game_id
LIMIT sqlc.arg('limit');

-- name: GetLastGameID :one
SELECT COALESCE(MAX(game_id), 0) AS last_game_id
FROM games;
