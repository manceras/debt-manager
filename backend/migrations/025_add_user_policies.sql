-- +goose Up
-- +goose StatementBegin
create schema if not exists app;

create or replace function app.current_user_id()
returns uuid
language sql stable as $$
  select nullif(current_setting('app.current_user', true), '')::uuid
$$;


create policy users_can_read_shared_list_members
on public.users
for select
using (
  exists (
    select 1
    from public.users_lists me
    join public.users_lists them
      on me.list_id = them.list_id
    where me.user_id  = app.current_user_id()
      and them.user_id = public.users.id
  )
);

create index if not exists users_lists_list_user_idx on public.users_lists (list_id, user_id);
create index if not exists users_lists_user_idx      on public.users_lists (user_id);

-- +goose StatementEnd
