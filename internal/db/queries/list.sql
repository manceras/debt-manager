-- name: CreateList :one
INSERT INTO lists (title, currency) VALUES ($1, $2) RETURNING *;
