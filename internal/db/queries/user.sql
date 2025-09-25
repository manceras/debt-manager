-- name: CreateUser :one
INSERT INTO users (email, password_hash, password_algo, username)
VALUES ($1, $2, $3, $4)
RETURNING id, email, username, created_at, password_changed_at, last_login_at;
