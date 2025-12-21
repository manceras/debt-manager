
-- +goose Up
-- +goose StatementBegin
DROP TRIGGER IF EXISTS users_lists_keep_one ON public.users_lists;
-- +goose StatementEnd
