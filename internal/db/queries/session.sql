
-- name: CreateSession :one
INSERT INTO app.sessions (user_id, expires_at, user_agent, ip) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM app.sessions WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO app.refresh_tokens (id, session_id, token_hash, expires_at, parent_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, session_id, created_at, expires_at;

-- name: AuthRefreshLookup :one
SELECT
  rt.id           AS rt_id,
  rt.session_id   AS session_id,
  rt.expires_at   AS rt_expires_at,
  rt.revoked_at   AS rt_revoked_at,
  rt.replaced_by_id AS rt_replaced_by_id,
  s.user_id       AS user_id,
  s.revoked_at    AS session_revoked_at,
  s.expires_at AS max_expires_at
FROM app.refresh_tokens rt
JOIN app.sessions s ON s.id = rt.session_id
WHERE rt.token_hash = $1
LIMIT 1;

-- name: MarkOldTokenReplaced :execrows
UPDATE app.refresh_tokens
SET replaced_by_id = $2
WHERE id = $1 AND replaced_by_id IS NULL;

-- name: RevokeWholeSession :exec
UPDATE app.sessions
SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL;

-- (optional, keep data tidy)
-- name: RevokeAllTokensInSession :exec
UPDATE app.refresh_tokens
SET revoked_at = now()
WHERE session_id = $1 AND revoked_at IS NULL;
