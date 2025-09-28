
-- +goose Up
-- +goose StatementBegin
-- Add the email column to the view app.login_secret
CREATE OR REPLACE VIEW app.login_secret AS
SELECT u.id, u.password_hash, u.password_algo, u.email
FROM public.users u;
-- +goose StatementEnd
