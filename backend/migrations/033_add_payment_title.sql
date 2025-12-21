-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.payments
  ADD COLUMN title TEXT;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.payments
  DROP COLUMN IF EXISTS title;
-- +goose StatementEnd
