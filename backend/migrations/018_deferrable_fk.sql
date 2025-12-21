
-- +goose Up
-- +goose StatementBegin
ALTER TABLE app.refresh_tokens
  DROP CONSTRAINT IF EXISTS refresh_tokens_replaced_by_id_fkey;

ALTER TABLE app.refresh_tokens
  ADD CONSTRAINT refresh_tokens_replaced_by_id_fkey
  FOREIGN KEY (replaced_by_id)
  REFERENCES app.refresh_tokens(id)
  ON DELETE SET NULL
  DEFERRABLE INITIALLY DEFERRED;

-- +goose StatementEnd
