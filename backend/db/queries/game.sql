
-- name: CreateGame :exec
INSERT INTO games (game_id, picks)
VALUES (?, ?);

-- name: GetGameByGameID :one
SELECT game_id, picks, created_at
FROM games
WHERE game_id = ?;

-- name: GetGameByLastGameID :one
SELECT game_id, picks, created_at
FROM games
WHERE game_id = (
    SELECT COALESCE(MAX(g.game_id), 0) - 1
    FROM games g
);

-- name: GetGamesByRange :many
SELECT game_id, picks, created_at
FROM games
WHERE game_id >= sqlc.arg('start')
ORDER BY game_id
LIMIT sqlc.arg('limit');

-- name: GetLastGameID :one
SELECT COALESCE(MAX(game_id), 0) + 0 AS last_game_id
FROM games;
