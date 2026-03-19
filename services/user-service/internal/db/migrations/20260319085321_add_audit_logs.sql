-- +goose Up
SELECT 'up SQL query';

-- Improve oauth table
ALTER TABLE oauths DROP COLUMN token;

-- create new table audit logs
CREATE TABLE audit_logs (
  id SERIAL PRIMARY KEY,
  account_id INT NOT NULL,
  action VARCHAR(255) NOT NULL,
  ip_address VARCHAR(255) NOT NULL,
  user_agent VARCHAR(255) NOT NULL,
  device_id VARCHAR(255) NOT NULL,
  metadata JSONB,
  timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES accounts(id)
);

-- +goose Down
SELECT 'down SQL query';
