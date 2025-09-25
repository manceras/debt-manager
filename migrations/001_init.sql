-- +goose UP
-- +goose StatementBegin
CREATE ROLE app_auth NOINHERIT;
CREATE ROLE app_admin NOINHERIT;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SCHEMA IF NOT EXISTS app;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS currency;
DROP SCHEMA IF EXISTS app CASCADE;
DROP ROLE IF EXISTS app_auth;
DROP ROLE IF EXISTS app_admin;
-- +goose StatementEnd
