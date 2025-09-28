
-- +goose Up
-- +goose StatementBegin
ALTER TABLE app.sessions
ALTER COLUMN id SET DEFAULT gen_random_uuid();
-- +goose StatementEnd
