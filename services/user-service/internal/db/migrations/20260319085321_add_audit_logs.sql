-- +goose Up
SELECT 'up SQL query';

-- Improve oauth table
ALTER TABLE oauths DROP COLUMN token;

-- create new table audit_logs
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  action VARCHAR(255) NOT NULL,
  ip_address VARCHAR(255) NOT NULL,
  user_agent VARCHAR(255) NOT NULL,
  device_id VARCHAR(255) NOT NULL,
  metadata JSONB,
  timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (account_id) REFERENCES accounts(id) -- Changed from user_id to account_id
);

-- +goose Down
SELECT 'down SQL query';

-- Drop the audit_logs table
DROP TABLE IF EXISTS audit_logs;
