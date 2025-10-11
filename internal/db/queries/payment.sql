-- name: CreatePayment :one
INSERT INTO public.payments (payer_user_id, amount, photo_url, list_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetAllPaymentsForList :many
SELECT * FROM public.payments
WHERE list_id = $1;

-- name: GetPaymentByID :one
SELECT * FROM public.payments
WHERE id = $1;

-- name: DeletePaymentByID :exec
DELETE FROM public.payments
WHERE id = $1;
