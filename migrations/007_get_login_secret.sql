-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.get_login_secret(_email text)
RETURNS TABLE (
  id uuid,
  password_hash text,
  password_algo text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
  RETURN QUERY
  SELECT u.id, u.password_hash, u.password_algo
  FROM public.users u
  WHERE u.email = _email;
END;
$$;

REVOKE ALL ON FUNCTION app.get_login_secret(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.get_login_secret(text) TO app_auth;
-- +goose StatementEnd
