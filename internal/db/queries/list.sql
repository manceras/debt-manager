-- name: CreateList :exec
INSERT INTO lists (id, title, currency) VALUES ($1, $2, $3);

-- name: CreateUserListRelation :one
INSERT INTO users_lists (user_id, list_id) VALUES ($1, $2) RETURNING *;

-- name: GetListByID :one
SELECT * FROM lists WHERE id = $1;
