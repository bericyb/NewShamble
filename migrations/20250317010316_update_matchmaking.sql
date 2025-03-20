-- +goose Up
-- +goose StatementBegin
ALTER TABLE quickplay_queue
    ADD COLUMN IF NOT EXISTS match_id UUID,
    ADD COLUMN IF NOT EXISTS matched_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
