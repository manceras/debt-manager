-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION app.set_user(id uuid)
RETURNS void 
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  PERFORM set_config('app.current_user', id::text, true);
END;
$$;

-- +goose StatementEnd
