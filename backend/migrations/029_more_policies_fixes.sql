-- +goose Up
-- +goose StatementBegin
/*
Purpose:
- Fix RLS bootstrap deadlock: after creating a list, no one can insert the first
  row in public.users_lists because existing INSERT policies require membership.
- Add a special INSERT policy that allows the caller to insert the FIRST membership
  for a list, but only for themselves.

Assumptions:
- RLS is enabled on public.users_lists (as in your earlier migrations).
- app.current_user_id() returns UUID of the session caller.
*/

-- Ensure RLS is on (harmless if already enabled)
ALTER TABLE public.users_lists ENABLE ROW LEVEL SECURITY;

-- Create the "first membership" policy if it does not already exist
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policy p
    WHERE p.polrelid = 'public.users_lists'::regclass
      AND p.polname  = 'users_lists_insert_initial'
  ) THEN
    CREATE POLICY users_lists_insert_initial
    ON public.users_lists
    FOR INSERT
    WITH CHECK (
      -- must be logged in
      app.current_user_id() IS NOT NULL
      -- can only add yourself as the first member
      AND user_id = app.current_user_id()
      -- only when the list has no members yet
      AND NOT EXISTS (
        SELECT 1
        FROM public.users_lists ul2
        WHERE ul2.list_id = users_lists.list_id
      )
    );
  END IF;
END
$$;

-- Optional hardening: make sure there is at least one policy that governs
-- additional inserts (inviting others) which requires membership.
-- We DO NOT overwrite your existing policy if one is present.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policy p
    WHERE p.polrelid = 'public.users_lists'::regclass
      AND p.polcmd   = 'r'  -- 'r' = INSERT
      AND p.polname  <> 'users_lists_insert_initial'
  ) THEN
    -- Create a conservative policy for adding members only if already a member.
    -- If you already have one, this block will be skipped.
    CREATE POLICY users_lists_insert_members
    ON public.users_lists
    FOR INSERT
    WITH CHECK (app.is_member(list_id));
  END IF;
END
$$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
-- Remove only what we added; leave your existing policies intact.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_policy p
    WHERE p.polrelid = 'public.users_lists'::regclass
      AND p.polname  = 'users_lists_insert_initial'
  ) THEN
    DROP POLICY users_lists_insert_initial ON public.users_lists;
  END IF;

  -- If we created the fallback "users_lists_insert_members" above, drop it as well.
  IF EXISTS (
    SELECT 1
    FROM pg_policy p
    WHERE p.polrelid = 'public.users_lists'::regclass
      AND p.polname  = 'users_lists_insert_members'
  ) THEN
    DROP POLICY users_lists_insert_members ON public.users_lists;
  END IF;
END
$$;
-- +goose StatementEnd
