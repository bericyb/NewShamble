-- +goose Up
-- +goose StatementBegin
ALTER TABLE quickplay_queue
ADD COLUMN status VARCHAR(20) DEFAULT 'pending' NOT NULL;

ALTER TABLE ranked_queue
ADD COLUMN status VARCHAR(20) DEFAULT 'pending' NOT NULL;

ALTER TABLE tournament_queue
ADD COLUMN status VARCHAR(20) DEFAULT 'pending' NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
