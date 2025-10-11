
-- name: CreateInvitation :one
INSERT INTO public.invitations (invited_to_list_id, expires_at, created_by, hash)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetAllInvitationsForList :many
SELECT * FROM public.invitations
WHERE invited_to_list_id = $1;

-- name: GetInvitationByHash :one
SELECT * FROM public.invitations
WHERE hash = $1;

-- name: RevokeInvitationByID :exec
UPDATE public.invitations
SET revoked_at = now()
WHERE id = $1;

-- name: GetInvitationByID :one
SELECT * FROM public.invitations
WHERE id = $1;
