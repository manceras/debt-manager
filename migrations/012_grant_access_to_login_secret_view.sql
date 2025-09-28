-- +goose Up
-- +goose StatementBegin
GRANT SELECT ON app.login_secret TO app_auth;
-- +goose StatementEnd
