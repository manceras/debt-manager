-- name: CreateList :exec
INSERT INTO lists (id, title, currency) VALUES ($1, $2, $3);

-- name: CreateUserListRelation :one
INSERT INTO users_lists (user_id, list_id) VALUES ($1, $2) RETURNING *;

-- name: GetListByID :one
SELECT * FROM lists WHERE id = $1;

-- name: UpdateList :exec
UPDATE lists SET title = $2, currency = $3 WHERE id = $1;

-- name: DeleteList :exec
DELETE FROM lists WHERE id = $1;

-- name: GetAllLists :many
SELECT * FROM lists;

-- name: GetUsersInList :many
SELECT id, username, email FROM users
JOIN users_lists ON user_id = id
WHERE list_id = $1 AND id <> app.current_user_id();
