-- +goose Up
-- +goose StatementBegin
-- Align getter with setter without breaking dependencies
CREATE OR REPLACE FUNCTION app.current_user_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $$
  SELECT NULLIF(current_setting('app.current_user', true), '')::uuid
$$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
-- Restore previous behavior (reading the old key) without dropping
CREATE OR REPLACE FUNCTION app.current_user_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $$
  SELECT NULLIF(current_setting('app.user_id', true), '')::uuid
$$;
-- +goose StatementEnd
