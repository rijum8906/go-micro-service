-- +goose Up
DROP TABLE IF EXISTS account_securities;

ALTER TABLE accounts 
  ADD COLUMN is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN email_verified_at TIMESTAMPTZ,
  ADD COLUMN two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN two_factor_enabled_at TIMESTAMPTZ,
  ADD COLUMN two_factor_secret TEXT;


-- +goose Down
