-- +goose Up
-- +goose StatementBegin
-- 1) Grant table privileges to the app role
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.invitations TO app_auth;

-- 2) (Re)enable RLS and add membership-based policies (safe if already present)
ALTER TABLE public.invitations ENABLE ROW LEVEL SECURITY;

-- Allow members of the list to see invitations
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policy
    WHERE polrelid='public.invitations'::regclass AND polname='invitations_select_members'
  ) THEN
    CREATE POLICY invitations_select_members
    ON public.invitations
    FOR SELECT
    USING ( app.is_member(invited_to_list_id) );
  END IF;
END$$;

-- Allow members to create invitations for their lists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policy
    WHERE polrelid='public.invitations'::regclass AND polname='invitations_insert_members'
  ) THEN
    CREATE POLICY invitations_insert_members
    ON public.invitations
    FOR INSERT
    WITH CHECK ( app.is_member(invited_to_list_id) );
  END IF;
END$$;

-- Allow members to update/delete invitations on their lists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policy
    WHERE polrelid='public.invitations'::regclass AND polname='invitations_update_members'
  ) THEN
    CREATE POLICY invitations_update_members
    ON public.invitations
    FOR UPDATE
    USING ( app.is_member(invited_to_list_id) )
    WITH CHECK ( app.is_member(invited_to_list_id) );
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_policy
    WHERE polrelid='public.invitations'::regclass AND polname='invitations_delete_members'
  ) THEN
    CREATE POLICY invitations_delete_members
    ON public.invitations
    FOR DELETE
    USING ( app.is_member(invited_to_list_id) );
  END IF;
END$$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLE public.invitations FROM app_auth;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_policy
             WHERE polrelid='public.invitations'::regclass AND polname='invitations_select_members')
  THEN
    DROP POLICY invitations_select_members ON public.invitations;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_policy
             WHERE polrelid='public.invitations'::regclass AND polname='invitations_insert_members')
  THEN
    DROP POLICY invitations_insert_members ON public.invitations;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_policy
             WHERE polrelid='public.invitations'::regclass AND polname='invitations_update_members')
  THEN
    DROP POLICY invitations_update_members ON public.invitations;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_policy
             WHERE polrelid='public.invitations'::regclass AND polname='invitations_delete_members')
  THEN
    DROP POLICY invitations_delete_members ON public.invitations;
  END IF;
END$$;

-- If we granted a sequence, revoke it (only if it exists)
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_class c
    JOIN pg_namespace n ON n.oid=c.relnamespace
    WHERE n.nspname='public' AND c.relkind='S' AND c.relname='invitations_id_seq'
  ) THEN
    REVOKE USAGE, SELECT, UPDATE ON SEQUENCE public.invitations_id_seq FROM app_auth;
  END IF;
END$$;
-- +goose StatementEnd
