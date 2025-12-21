-- +goose Up
-- +goose StatementBegin
-- Ensure schema exists
CREATE SCHEMA IF NOT EXISTS app;

-- Columns (fix name + defaults)
ALTER TABLE public.users
  ADD COLUMN IF NOT EXISTS password_hash       text,
  ADD COLUMN IF NOT EXISTS password_algo       text NOT NULL DEFAULT 'argon2id',
  ADD COLUMN IF NOT EXISTS password_changed_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS last_login_at       timestamptz;  -- NULL until first login

-- Optional sanity: constrain known algos
-- ALTER TABLE public.users ADD CONSTRAINT users_password_algo_chk
--   CHECK (password_algo IN ('argon2id','bcrypt'));

-- Case-insensitive unique email
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uidx
  ON public.users (lower(email));

-- Lock down raw table
REVOKE ALL ON TABLE public.users FROM PUBLIC;
REVOKE ALL ON TABLE public.users FROM app_auth;

-- Allow writes for app_auth (no reads)
GRANT INSERT ON public.users TO app_auth;
GRANT UPDATE (last_login_at) ON public.users TO app_auth;  -- narrow update

-- ===== RLS POLICIES (assumes RLS already enabled on users) =====
-- Self-select so INSERT ... RETURNING and view work for own row
CREATE POLICY users_select_self
ON public.users
FOR SELECT TO app_auth
USING (id::text = current_setting('app.current_user', true));

-- Admin can see all
CREATE POLICY users_select_all_admin
ON public.users
FOR SELECT TO app_admin
USING (true);

-- Allow app_auth to insert any row (shape enforced by constraints)
CREATE POLICY users_insert_any
ON public.users
FOR INSERT TO app_auth
WITH CHECK (true);

-- Allow app_auth to update only their own row (column scope handled by GRANT)
CREATE POLICY users_update_self
ON public.users
FOR UPDATE TO app_auth
USING (id::text = current_setting('app.current_user', true))
WITH CHECK (id::text = current_setting('app.current_user', true));

-- If you are using FORCE ROW LEVEL SECURITY, also allow the function owner (e.g., app_admin)
-- to bypass restrictions within SECURITY DEFINER functions:
-- CREATE POLICY users_all_admin
-- ON public.users
-- FOR ALL TO app_admin
-- USING (true) WITH CHECK (true);

-- ===== Safe read-only view =====
CREATE OR REPLACE VIEW app.users_safe AS
SELECT id, username, email, created_at, password_changed_at, last_login_at
FROM public.users;

GRANT USAGE ON SCHEMA app TO app_auth;
GRANT SELECT ON app.users_safe TO app_auth;

-- ===== Functions (SECURITY DEFINER) =====

-- Register user
CREATE OR REPLACE FUNCTION app.register_user(
  _email          text,
  _username       text,
  _password_hash  text,
  _password_algo  text DEFAULT 'argon2id'
) RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE _id uuid;
BEGIN
  INSERT INTO public.users (id, email, username, password_hash, password_algo)
  VALUES (gen_random_uuid(), _email, _username, _password_hash, _password_algo)
  RETURNING id INTO _id;
  RETURN _id;
EXCEPTION
  WHEN unique_violation THEN
    RAISE EXCEPTION 'email_already_registered' USING ERRCODE = '23505';
END;
$$;

REVOKE ALL ON FUNCTION app.register_user(text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.register_user(text, text, text, text) TO app_auth;

-- Change password (done via function; app_auth has no direct UPDATE on hash)
CREATE OR REPLACE FUNCTION app.update_user_password(
  _user_id           uuid,
  _new_password_hash text,
  _new_password_algo text DEFAULT 'argon2id'
) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
  UPDATE public.users
  SET password_hash = _new_password_hash,
      password_algo = _new_password_algo,
      password_changed_at = now()
  WHERE id = _user_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'user_not_found' USING ERRCODE = '02000';
  END IF;
END;
$$;

REVOKE ALL ON FUNCTION app.update_user_password(uuid, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.update_user_password(uuid, text, text) TO app_auth;

-- Update last login (can also be done directly due to GRANT UPDATE (last_login_at))
CREATE OR REPLACE FUNCTION app.update_last_login(_user_id uuid) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
  UPDATE public.users
  SET last_login_at = now()
  WHERE id = _user_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'user_not_found' USING ERRCODE = '02000';
  END IF;
END;
$$;

REVOKE ALL ON FUNCTION app.update_last_login(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.update_last_login(uuid) TO app_auth;
-- +goose StatementEnd
