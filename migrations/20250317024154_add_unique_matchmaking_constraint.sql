-- +goose Up
-- +goose StatementBegin
ALTER TABLE quickplay_queue
ADD CONSTRAINT unique_quickplay_matchmaking_constraint UNIQUE (user_id);

ALTER TABLE ranked_queue
ADD CONSTRAINT unique_ranked_matchmaking_constraint UNIQUE (user_id);

ALTER TABLE tournament_queue
ADD CONSTRAINT unique_tournament_matchmaking_constraint UNIQUE (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
