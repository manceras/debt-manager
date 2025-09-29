-- +goose Up
-- +goose StatementBegin

-- 1) Safer setter (no result expected in PL/pgSQL)
CREATE OR REPLACE FUNCTION app.set_user(id uuid)
RETURNS void 
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  PERFORM set_config('app.current_user', id::text, true);
END;
$$;

-- 2) Helper: current user id from GUC
CREATE OR REPLACE FUNCTION app.current_user_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $$
  SELECT current_setting('app.current_user', true)::uuid
$$;

-- 3) Helper: membership check (SECURITY DEFINER so it bypasses RLS)
--    IMPORTANT: this function should be OWNED by the same role that owns public.users_lists
--    so that it can see rows regardless of RLS while evaluating policies.
CREATE OR REPLACE FUNCTION app.is_member(_list_id uuid)
RETURNS boolean
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public, app
AS $$
  SELECT EXISTS (
    SELECT 1
    FROM public.users_lists ul
    WHERE ul.list_id = _list_id
      AND ul.user_id = app.current_user_id()
  )
$$;

-- 4) Ensure RLS is on (idempotent)
ALTER TABLE public.users       ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.lists       ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.users_lists ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.payments    ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.divisions   ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.categories  ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.payments_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.deposits    ENABLE ROW LEVEL SECURITY;

-- 5) Drop recursive/problematic policies (if they exist)
DROP POLICY IF EXISTS users_lists_read            ON public.users_lists;
DROP POLICY IF EXISTS users_lists_write           ON public.users_lists;
DROP POLICY IF EXISTS users_lists_delete          ON public.users_lists;
DROP POLICY IF EXISTS users_lists_delete_keep_one ON public.users_lists;

DROP POLICY IF EXISTS users_is_self      ON public.users;
DROP POLICY IF EXISTS lists_members_only ON public.lists;

DROP POLICY IF EXISTS payments_members_only             ON public.payments;
DROP POLICY IF EXISTS divisions_members_only            ON public.divisions;
DROP POLICY IF EXISTS categories_read_all               ON public.categories;
DROP POLICY IF EXISTS payments_categories_members_only  ON public.payments_categories;
DROP POLICY IF EXISTS deposits_members_only             ON public.deposits;

-- 6) Recreate policies using helpers (no self-reference on users_lists)

-- Users: can only see/modify self
CREATE POLICY users_is_self ON public.users
  USING (id = app.current_user_id())
  WITH CHECK (id = app.current_user_id());

-- users_lists: 
--   - read: own rows OR membership to that list (via helper, no self-select)
--   - insert: can only add self
--   - delete: only if member (no "keep one" here; we handle that via a trigger below)
CREATE POLICY users_lists_read ON public.users_lists
  FOR SELECT
  USING (
    user_id = app.current_user_id()
    OR app.is_member(list_id)
  );

CREATE POLICY users_lists_write ON public.users_lists
  FOR INSERT
  WITH CHECK (user_id = app.current_user_id());

CREATE POLICY users_lists_delete ON public.users_lists
  FOR DELETE
  USING (app.is_member(list_id));

-- lists: members only
CREATE POLICY lists_members_only ON public.lists
  USING (app.is_member(id))
  WITH CHECK (app.is_member(id));

-- payments: members of the payment's list
CREATE POLICY payments_members_only ON public.payments
  USING (app.is_member(list_id))
  WITH CHECK (app.is_member(list_id));

-- divisions: members of parent payment's list
CREATE POLICY divisions_members_only ON public.divisions
  USING (
    EXISTS (
      SELECT 1
      FROM public.payments p
      WHERE p.id = public.divisions.payment_id
        AND app.is_member(p.list_id)
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1
      FROM public.payments p
      WHERE p.id = public.divisions.payment_id
        AND app.is_member(p.list_id)
    )
  );

-- categories: global read (unchanged)
CREATE POLICY categories_read_all ON public.categories
  FOR SELECT
  USING (true);

-- payments_categories: members of the payment's list
CREATE POLICY payments_categories_members_only ON public.payments_categories
  USING (
    EXISTS (
      SELECT 1
      FROM public.payments p
      WHERE p.id = public.payments_categories.payment_id
        AND app.is_member(p.list_id)
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1
      FROM public.payments p
      WHERE p.id = public.payments_categories.payment_id
        AND app.is_member(p.list_id)
    )
  );

-- deposits: members only
CREATE POLICY deposits_members_only ON public.deposits
  USING (app.is_member(list_id))
  WITH CHECK (app.is_member(list_id));

-- 7) Prevent deleting the last membership of a list (no recursion).
--    Use a DEFERRABLE CONSTRAINT TRIGGER with a SECURITY DEFINER function.
CREATE OR REPLACE FUNCTION app.prevent_last_membership()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public, app
AS $$
DECLARE
  remaining bigint;
BEGIN
  -- AFTER DELETE: the row is gone, check how many remain for that list
  SELECT COUNT(*) INTO remaining
  FROM public.users_lists
  WHERE list_id = OLD.list_id;

  IF remaining = 0 THEN
    RAISE EXCEPTION 'Cannot delete the last membership for list %', OLD.list_id
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NULL;
END;
$$;

-- Drop the trigger if exists to avoid duplicates
DROP TRIGGER IF EXISTS users_lists_keep_one ON public.users_lists;

-- Create as DEFERRABLE so multi-row deletes can be validated at COMMIT
CREATE CONSTRAINT TRIGGER users_lists_keep_one
AFTER DELETE ON public.users_lists
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION app.prevent_last_membership();

-- Privileges: (RLS still applies at query time)
GRANT USAGE ON SCHEMA app, public TO app_auth, app_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_auth, app_admin;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- Drop trigger and helper functions
DROP TRIGGER IF EXISTS users_lists_keep_one ON public.users_lists;
DROP FUNCTION IF EXISTS app.prevent_last_membership();

-- Drop rewritten policies
DROP POLICY IF EXISTS deposits_members_only             ON public.deposits;
DROP POLICY IF EXISTS payments_categories_members_only  ON public.payments_categories;
DROP POLICY IF EXISTS categories_read_all               ON public.categories;
DROP POLICY IF EXISTS divisions_members_only            ON public.divisions;
DROP POLICY IF EXISTS payments_members_only             ON public.payments;
DROP POLICY IF EXISTS users_lists_delete                ON public.users_lists;
DROP POLICY IF EXISTS users_lists_write                 ON public.users_lists;
DROP POLICY IF EXISTS users_lists_read                  ON public.users_lists;
DROP POLICY IF EXISTS users_is_self                     ON public.users;
DROP POLICY IF EXISTS lists_members_only                ON public.lists;

-- Drop helpers
DROP FUNCTION IF EXISTS app.is_member(uuid);
DROP FUNCTION IF EXISTS app.current_user_id();

-- Restore original policies (as provided earlier)

CREATE POLICY users_is_self ON public.users
  USING (id::text = current_setting('app.current_user', true))
  WITH CHECK (id::text = current_setting('app.current_user', true));

CREATE POLICY users_lists_read ON public.users_lists
  USING (
    user_id::text = current_setting('app.current_user', true)
    OR EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.users_lists.list_id
    )
  );

CREATE POLICY lists_members_only ON public.lists
  USING (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.lists.id
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.lists.id
    )
  );

CREATE POLICY users_lists_write ON public.users_lists
  FOR INSERT WITH CHECK (user_id::text = current_setting('app.current_user', true));

CREATE POLICY users_lists_delete ON public.users_lists
  FOR DELETE
  USING (
    EXISTS (
      SELECT 1
      FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.users_lists.list_id
    )
  );

CREATE POLICY payments_members_only ON public.payments
  USING (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.payments.list_id
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.payments.list_id
    )
  );

CREATE POLICY divisions_members_only ON public.divisions
  USING (
    EXISTS (
      SELECT 1 FROM public.payments p
      JOIN public.users_lists ul ON ul.list_id = p.list_id
      WHERE p.id = public.divisions.payment_id
        AND ul.user_id::text = current_setting('app.current_user', true)
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.payments p
      JOIN public.users_lists ul ON ul.list_id = p.list_id
      WHERE p.id = public.divisions.payment_id
        AND ul.user_id::text = current_setting('app.current_user', true)
    )
  );

CREATE POLICY categories_read_all ON public.categories
  FOR SELECT USING (true);

CREATE POLICY payments_categories_members_only ON public.payments_categories
  USING (
    EXISTS (
      SELECT 1
      FROM public.payments p
      JOIN public.users_lists ul ON ul.list_id = p.list_id
      WHERE p.id = public.payments_categories.payment_id
        AND ul.user_id::text = current_setting('app.current_user', true)
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1
      FROM public.payments p
      JOIN public.users_lists ul ON ul.list_id = p.list_id
      WHERE p.id = public.payments_categories.payment_id
        AND ul.user_id::text = current_setting('app.current_user', true)
    )
  );

CREATE POLICY deposits_members_only ON public.deposits
  USING (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.deposits.list_id
    )
  )
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.deposits.list_id
    )
  );

-- Keep set_user as last defined version (or restore original if desired)
CREATE OR REPLACE FUNCTION app.set_user(id uuid)
RETURNS void 
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  SELECT set_config('app.current_user', id::text, true);
END;
$$;

-- +goose StatementEnd
