
-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.prevent_last_membership()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public, app
AS $$
DECLARE
  list_exists boolean;
  remaining   bigint;
BEGIN
  -- If the parent list is gone (or being deleted in this tx), don't enforce
  SELECT EXISTS (SELECT 1 FROM public.lists WHERE id = OLD.list_id)
    INTO list_exists;

  IF NOT list_exists THEN
    RETURN NULL; -- allow deletions that come from list deletion
  END IF;

  -- Enforce "at least one member" only for lists that still exist
  SELECT COUNT(*) INTO remaining
  FROM public.users_lists
  WHERE list_id = OLD.list_id;

  IF remaining = 0 THEN
    RAISE EXCEPTION 'Cannot delete the last membership for list %', OLD.list_id
      USING ERRCODE = '23514'; -- check_violation
  END IF;

  RETURN NULL;
END;
$$;

-- (Re)create the trigger as deferrable so multi-row ops validate at COMMIT
DROP TRIGGER IF EXISTS users_lists_keep_one ON public.users_lists;

CREATE CONSTRAINT TRIGGER users_lists_keep_one
AFTER DELETE ON public.users_lists
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION app.prevent_last_membership();

-- +goose StatementEnd
