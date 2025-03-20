-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS quickplay_queue (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    queued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ranked_queue (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    elo INTEGER NOT NULL,
    queued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournament_queue (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    tournament_id INTEGER NOT NULL,
    level int NOT NULL,
    queued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
