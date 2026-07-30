-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_token (
    selector VARCHAR (128) PRIMARY KEY,
    private_hash BYTEA NOT NULL,  
    user_id INT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
    issued_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    expired_at TIMESTAMP NOT NULL,
    revoked bool NOT NULL DEFAULT FALSE)
-- +goose StatementEnd