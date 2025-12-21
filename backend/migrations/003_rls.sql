-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.set_user(id uuid)
RETURNS void 
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
	SELECT set_config('app.current_user', id::text, true);
END;
$$;

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE lists ENABLE ROW LEVEL SECURITY;
ALTER TABLE users_lists ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE divisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE deposits ENABLE ROW LEVEL SECURITY;

CREATE OR REPLACE VIEW app.v_memberships AS
SELECT ul.user_id, ul.list_id
FROM public.users_lists ul;

-- POLICIES

-- Users can only see and modify their own user record
CREATE POLICY users_is_self ON users
  USING (id::text = current_setting('app.current_user', true))
  WITH CHECK (id::text = current_setting('app.current_user', true));

-- Users can only see and modify their own lists
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

-- Users can only add lists for themselves
CREATE POLICY users_lists_write ON public.users_lists
  FOR INSERT WITH CHECK (user_id::text = current_setting('app.current_user', true));

-- Users can only delete their own lists or lists they belong to
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

-- Payments: only members of the payment’s list
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

-- Divisions: access allowed if you can access the parent payment (and thus the list)
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

-- Categories: global read; restrict write to admins if you want (for now: read-all)
CREATE POLICY categories_read_all ON public.categories
  FOR SELECT USING (true);

-- Payments↔Categories: members of the payment’s list
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

-- Deposits: members of the list only
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

-- Prevent users from deleting their last membership (to avoid orphaned lists)
CREATE POLICY users_lists_delete_keep_one ON public.users_lists
  FOR DELETE USING (
    EXISTS (
      SELECT 1 FROM public.users_lists ul
      WHERE ul.user_id::text = current_setting('app.current_user', true)
        AND ul.list_id = public.users_lists.list_id
    )
    AND (
      SELECT COUNT(*) FROM public.users_lists x
      WHERE x.list_id = public.users_lists.list_id
    ) > 1
  );

-- Privileges (RLS still applies)
GRANT USAGE ON SCHEMA app, public TO app_auth, app_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_auth, app_admin;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS deposits_members_only      ON public.deposits;
DROP POLICY IF EXISTS payments_categories_members_only ON public.payments_categories;
DROP POLICY IF EXISTS categories_read_all        ON public.categories;
DROP POLICY IF EXISTS divisions_members_only     ON public.divisions;
DROP POLICY IF EXISTS payments_members_only      ON public.payments;
DROP POLICY IF EXISTS users_lists_delete         ON public.users_lists;
DROP POLICY IF EXISTS users_lists_write          ON public.users_lists;
DROP POLICY IF EXISTS users_lists_read           ON public.users_lists;
DROP POLICY IF EXISTS users_is_self              ON public.users;
DROP POLICY IF EXISTS lists_members_only         ON public.lists;
-- +goose StatementEnd


