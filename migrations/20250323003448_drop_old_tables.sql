-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS tournament_queue CASCADE;
DROP TABLE IF EXISTS ranked_queue CASCADE;
DROP TABLE IF EXISTS quickplay_queue CASCADE;
DROP TABLE IF EXISTS ranked_matchmaking_queue CASCADE;
DROP TABLE IF EXISTS matchmaking_matches CASCADE;
DROP TABLE IF EXISTS casual_matchmaking_queue CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
