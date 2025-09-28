
-- +goose Up
-- +goose StatementBegin
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE app.sessions TO app_auth;
-- +goose StatementEnd
