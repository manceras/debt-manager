
-- +goose Up
-- +goose StatementBegin
-- keep existing lists_members_only for SELECT/UPDATE/DELETE
DROP POLICY IF EXISTS lists_members_only ON public.lists;
CREATE POLICY lists_members_only ON public.lists
  USING (app.is_member(id))
  WITH CHECK (app.is_member(id));  -- applies to UPDATE only now

-- add INSERT policy that doesn't require prior membership
CREATE POLICY lists_insert ON public.lists
  FOR INSERT
  WITH CHECK (app.current_user_id() IS NOT NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS lists_insert ON public.lists;
DROP POLICY IF EXISTS lists_members_only ON public.lists;
-- +goose StatementEnd
