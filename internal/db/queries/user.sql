-- name: CreateUser :one
SELECT * FROM app.register_user($1, $2, $3, $4);

-- name: GetUserByID :one
SELECT * FROM app.users_safe WHERE id = $1;
