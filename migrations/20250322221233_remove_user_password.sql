-- +goose Up
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN password;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
