-- name: CreateUser :one
SELECT * FROM app.register_user($1, $2, $3, $4);

-- name: GetUserByID :one
SELECT * FROM app.users_safe WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM app.users_safe WHERE email = $1;

-- name: GetLoginSecretsByEmail :one
SELECT * FROM app.login_secret WHERE email = $1;

-- name: UpdateUserLastLogin :exec
SELECT app.update_last_login($1);

-- name: GetUsersFromList :many
SELECT u.* FROM app.users_safe u
JOIN users_lists ul ON u.id = ul.user_id
WHERE ul.list_id = $1;
