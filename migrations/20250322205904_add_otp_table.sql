-- +goose Up
-- +goose StatementBegin
CREATE TABLE otp (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(15) NOT NULL,
    code VARCHAR(6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users
ADD COLUMN new_tournaments_notif BOOLEAN DEFAULT TRUE,
ADD COLUMN friends_joined_notif BOOLEAN DEFAULT TRUE,
ADD COLUMN tournament_starting_notif BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE otp;
-- +goose StatementEnd
