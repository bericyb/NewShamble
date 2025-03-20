-- +goose Up 
CREATE TABLE ranked_matchmaking_queue (
    queue_id SERIAL PRIMARY KEY,
    player_id INT UNIQUE,
    skill_rating INT NOT NULL,
    queue_time TIMESTAMP DEFAULT NOW ()
);

CREATE TABLE casual_matchmaking_queue (
    queue_id SERIAL PRIMARY KEY,
    player_id INT UNIQUE,
    queue_time TIMESTAMP DEFAULT NOW ()
);

CREATE TABLE matchmaking_matches (
    match_id SERIAL PRIMARY KEY,
    player1_id UUID,
    player2_id UUID,
    player1_ready BOOLEAN DEFAULT FALSE,
    player2_ready BOOLEAN DEFAULT FALSE,
    match_time TIMESTAMP DEFAULT NOW ()
);
