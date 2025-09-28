
-- name: CreateSession :one
INSERT INTO app.sessions (user_id, expires_at, user_agent, ip) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM app.sessions WHERE id = $1;
