
-- +goose Up
-- +goose StatementBegin
GRANT UPDATE ON TABLE app.refresh_tokens TO app_auth;
-- +goose StatementEnd
