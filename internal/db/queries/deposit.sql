-- name: CreateDeposit :one
INSERT INTO deposits (amount, payer_user_id, payee_user_id, list_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetAllDepositsForListID :many
SELECT * FROM deposits WHERE list_id = $1;
