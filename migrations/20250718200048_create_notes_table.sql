-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE note (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  account_id INT NOT NULL,                           
  title VARCHAR(255) NOT NULL ,
  body TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP)
-- +goose StatementEnd




