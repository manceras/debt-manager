-- +goose Up
-- +goose StatementBegin
ALTER TABLE app.sessions
		ALTER COLUMN ip TYPE text USING ip::text;
-- +goose StatementEnd
