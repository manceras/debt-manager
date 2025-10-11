
-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.invitations (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	hash text UNIQUE NOT NULL,
	invited_to_list_id uuid REFERENCES lists(id) ON DELETE CASCADE,
	expires_at timestamptz NOT NULL,
	revoked_at timestamptz,
	created_at timestamptz DEFAULT now(),
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	used_by uuid REFERENCES users(id) ON DELETE SET NULL,
	used_at timestamptz
);

CREATE INDEX ON public.invitations (expires_at);
CREATE INDEX ON public.invitations (invited_to_list_id);
CREATE INDEX ON public.invitations (created_by);
CREATE INDEX ON public.invitations (used_by);

alter table public.invitations enable row level security;

-- Helper: current app user (GUC pattern if not on Supabase)
create or replace function app.current_user_id()
returns uuid language sql stable as $$
  select nullif(current_setting('app.user_id', true), '')::uuid
$$;

-- Policy: members can see/manage invitations of their lists
create policy invitations_member_select
on public.invitations
for select
using (
  exists (
    select 1 from public.users_lists ul
    where ul.list_id = invitations.invited_to_list_id
      and ul.user_id = app.current_user_id()
  )
);

create policy invitations_member_insert
on public.invitations
for insert
with check (
  exists (
    select 1 from public.users_lists ul
    where ul.list_id = invited_to_list_id
      and ul.user_id = app.current_user_id()
  ) and created_by = app.current_user_id()
);

create policy invitations_member_update
on public.invitations
for update
using (
  created_by = app.current_user_id() or
  exists (
    select 1 from public.users_lists ul
    where ul.list_id = invitations.invited_to_list_id
      and ul.user_id = app.current_user_id()
  )
);


create or replace function app.invitation_preview(p_hash text)
returns table(list_id uuid, list_title text, expires_at timestamptz, revoked_at timestamptz, used_at timestamptz)
language sql
security definer
set search_path = pg_catalog, public, app
as $$
  select i.invited_to_list_id, l.title, i.expires_at, i.revoked_at, i.used_at
  from public.invitations i
  join public.lists l on l.id = i.invited_to_list_id
  where i.hash = p_hash
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the preview function first (depends on invitations)
DROP FUNCTION IF EXISTS app.invitation_preview(text);

-- Drop RLS policies (will also be removed by DROP TABLE, but explicit is clearer)
DROP POLICY IF EXISTS invitations_member_update ON public.invitations;
DROP POLICY IF EXISTS invitations_member_insert ON public.invitations;
DROP POLICY IF EXISTS invitations_member_select ON public.invitations;

-- Drop the table (indexes and RLS settings go with it)
DROP TABLE IF EXISTS public.invitations;

-- If this function was introduced in this migration and not used elsewhere, drop it too.
DROP FUNCTION IF EXISTS app.current_user_id();

-- +goose StatementEnd
