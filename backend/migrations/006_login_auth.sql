-- +goose Up
-- +goose StatementBegin
ALTER ROLE app_auth WITH LOGIN PASSWORD 'secret';
-- +goose StatementEnd
