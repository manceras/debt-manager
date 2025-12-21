-- +goose Up
-- +goose StatementBegin
ALTER ROLE app_auth WITH PASSWORD 'secret';
-- +goose StatementEnd

