-- name: CreateList :exec
INSERT INTO lists (id, title, currency) VALUES ($1, $2, $3);

-- name: CreateUserListRelation :one
INSERT INTO users_lists (user_id, list_id) VALUES ($1, $2) RETURNING *;

-- name: GetListByID :one
SELECT * FROM lists WHERE id = $1;

-- name: UpdateList :exec
UPDATE lists
SET
  title    = COALESCE(sqlc.narg(title), title),
  currency = COALESCE(sqlc.narg(currency), currency)
WHERE id = $1;

-- name: DeleteList :one
DELETE FROM lists WHERE id = $1 RETURNING *;

-- name: GetAllLists :many
SELECT * FROM lists;

-- name: GetUsersInList :many
SELECT id, username, email FROM users
JOIN users_lists ON user_id = id
WHERE list_id = $1 AND id <> app.current_user_id();
