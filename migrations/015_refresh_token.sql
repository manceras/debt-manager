
-- +goose Up
-- +goose StatementBegin
CREATE TABLE app.refresh_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id uuid NOT NULL REFERENCES app.sessions(id) ON DELETE CASCADE,
  token_hash bytea NOT NULL UNIQUE, -- sha256(raw)
  parent_id uuid NULL REFERENCES app.refresh_tokens(id) ON DELETE SET NULL,
  replaced_by_id uuid NULL REFERENCES app.refresh_tokens(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz NULL
);

-- Exactly one active leaf per session (optional but handy)
CREATE UNIQUE INDEX refresh_tokens_one_active_leaf_per_session
  ON app.refresh_tokens (session_id)
  WHERE replaced_by_id IS NULL AND revoked_at IS NULL;


GRANT SELECT, INSERT ON app.refresh_tokens TO app_auth;

-- +goose StatementEnd
