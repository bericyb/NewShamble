-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION delete_old_quickplay_queue() RETURNS void AS $$
BEGIN
    DELETE FROM quickplay_queue WHERE queued_at < NOW() - INTERVAL '5 minutes';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION delete_old_ranked_queue() RETURNS void AS $$
BEGIN
    DELETE FROM ranked_queue WHERE queued_at < NOW() - INTERVAL '5 minutes';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION delete_old_tournament_queue() RETURNS void AS $$
BEGIN
    DELETE FROM tournament_queue WHERE queued_at < NOW() - INTERVAL '5 minutes';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION find_and_match_players() RETURNS VOID AS $$
DECLARE
    player RECORD;
    match RECORD;
    v_match_uuid UUID;
BEGIN
    FOR player IN SELECT * FROM quickplay_queue ORDER BY queued_at LOOP
		v_match_uuid := gen_random_uuid();
        SELECT * INTO match FROM quickplay_queue
        WHERE match_id = NULL and user_id != player.user_id
        ORDER BY queued_at ASC
        LIMIT 1;

        IF FOUND THEN
            UPDATE quickplay_queue SET match_id = v_match_uuid, matched_at = NOW() WHERE user_id IN (player.user_id, match.user_id);
        END IF;
    END LOOP;

    DELETE FROM quickplay_queue WHERE match_id IS NOT NULL AND matched_at < NOW() - INTERVAL '20 seconds';
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
