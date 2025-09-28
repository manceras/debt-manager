
-- +goose Up
-- +goose StatementBegin
CREATE TABLE app.sessions (
	id uuid PRIMARY KEY,
	user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
	created_at timestamptz DEFAULT now(),
	expires_at timestamptz NOT NULL,
	revoked_at timestamptz,
	user_agent text,
	ip inet
);

CREATE INDEX ON app.sessions (user_id);
CREATE INDEX ON app.sessions (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
