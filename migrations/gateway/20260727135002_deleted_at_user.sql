-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_user
ADD COLUMN deleted_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_user
DROP COLUMN deleted_at; 
-- +goose StatementEnd
