-- +goose Up 
ALTER TABLE users
DROP COLUMN full_name;
