-- +goose Up
-- +goose StatementBegin

-- Alter the app.refresh_tokens table to change the parent_id column to be a nullable UUID type and FK

ALTER TABLE app.refresh_tokens
		ALTER COLUMN parent_id DROP NOT NULL,
		ALTER COLUMN parent_id TYPE UUID USING parent_id::UUID,
		DROP CONSTRAINT refresh_tokens_parent_id_fkey,
		ADD CONSTRAINT refresh_tokens_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES app.refresh_tokens(id) ON DELETE CASCADE;

-- +goose StatementEnd
