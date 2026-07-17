-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_user ( 
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  login VARCHAR(255) NOT NULL,
  password VARCHAR(60) NOT NULL,
  email VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX ux_users_email_nocase ON app_user (LOWER(email));
CREATE UNIQUE INDEX ux_users_login_nocase ON app_user (LOWER(login));
-- +goose StatementEnd
