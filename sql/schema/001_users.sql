-- +goose Up
CREATE TABLE users (
  id uuid PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email TEXT NOT NULL UNIQUE,
  hashed_password TEXT NOT NULL DEFAULT 'notset'
);

-- +goose Down
DROP TABLE users;
