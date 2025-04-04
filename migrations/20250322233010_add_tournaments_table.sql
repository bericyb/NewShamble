-- +goose Up
-- +goose StatementBegin
CREATE TABLE tournaments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    status VARCHAR(50),
    emoji VARCHAR(10),
    prize VARCHAR(225) NOT NULL,
    prize_url VARCHAR(255),
    invite_level INT DEFAULT 0, 
    -- Default start time of 30 minutes from now
    start_date TIMESTAMP DEFAULT (NOW() + INTERVAL '30 minutes'),
    location VARCHAR(255),
    winner_id UUID DEFAULT NULL
);

ALTER TABLE users ADD COLUMN invite_level INT DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
