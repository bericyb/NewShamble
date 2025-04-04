-- +goose Up
-- +goose StatementBegin
ALTER TABLE otp
ADD CONSTRAINT unique_phone UNIQUE (phone, code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
