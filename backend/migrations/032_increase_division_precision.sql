-- +goose Up
ALTER TABLE public.divisions ALTER COLUMN amount TYPE NUMERIC(20,6);

-- +goose Down
ALTER TABLE public.divisions ALTER COLUMN amount TYPE NUMERIC(12,2);
