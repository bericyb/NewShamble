-- +goose Up
-- +goose StatementBegin
ALTER TABLE otp
ADD CONSTRAINT unique_only_phone UNIQUE (phone);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
