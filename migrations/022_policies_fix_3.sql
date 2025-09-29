
-- +goose Up
-- +goose StatementBegin
-- Clean up any existing policies to avoid name collisions
DROP POLICY IF EXISTS lists_members_only   ON public.lists;
DROP POLICY IF EXISTS lists_members_select ON public.lists;
DROP POLICY IF EXISTS lists_members_update ON public.lists;
DROP POLICY IF EXISTS lists_members_delete ON public.lists;
DROP POLICY IF EXISTS lists_insert         ON public.lists;

-- Members can SELECT lists they belong to
CREATE POLICY lists_members_select ON public.lists
  FOR SELECT
  USING (app.is_member(id));

-- Members can UPDATE lists they belong to
CREATE POLICY lists_members_update ON public.lists
  FOR UPDATE
  USING (app.is_member(id))
  WITH CHECK (app.is_member(id));

-- Members can DELETE lists they belong to
CREATE POLICY lists_members_delete ON public.lists
  FOR DELETE
  USING (app.is_member(id));

-- Anyone with a current user set can INSERT a new list
-- (ensure you add the creator to users_lists in the same transaction)
CREATE POLICY lists_insert ON public.lists
  FOR INSERT
  WITH CHECK (current_setting('app.current_user', true) IS NOT NULL);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS lists_insert         ON public.lists;
DROP POLICY IF EXISTS lists_members_delete ON public.lists;
DROP POLICY IF EXISTS lists_members_update ON public.lists;
DROP POLICY IF EXISTS lists_members_select ON public.lists;

-- Optional: recreate a single all-commands policy (not recommended)
CREATE POLICY lists_members_only ON public.lists
  USING (app.is_member(id))
  WITH CHECK (app.is_member(id));
-- +goose StatementEnd
