-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.divisions
  ALTER COLUMN amount TYPE NUMERIC(12,2)
  USING ROUND(amount, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.divisions
  ALTER COLUMN amount TYPE NUMERIC(20,6);
-- +goose StatementEnd
