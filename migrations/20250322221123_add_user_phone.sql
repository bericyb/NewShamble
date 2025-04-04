-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN phone VARCHAR(16);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
