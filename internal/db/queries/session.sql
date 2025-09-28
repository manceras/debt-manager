
-- name: CreateSession :one
INSERT INTO app.sessions (user_id, expires_at, user_agent, ip) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM app.sessions WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO app.refresh_tokens (session_id, token_hash, expires_at, parent_id)
VALUES ($1, $2, $3, $4)
RETURNING id, session_id, created_at, expires_at;
