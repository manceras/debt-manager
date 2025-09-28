
-- +goose Up
-- +goose StatementBegin
-- Delete the function app.get_login_secret if it exists
DROP FUNCTION IF EXISTS app.get_login_secret(text);

-- Create or replace the view app.login_secret
CREATE OR REPLACE VIEW app.login_secret AS
SELECT u.id, u.password_hash, u.password_algo
FROM public.users u;
-- +goose StatementEnd
