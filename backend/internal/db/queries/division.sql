
-- name: CreateDivision :one
INSERT INTO public.divisions (owe_user_id, amount, payment_id)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetDivisionsByPaymentID :many
SELECT * FROM public.divisions WHERE payment_id = $1;
