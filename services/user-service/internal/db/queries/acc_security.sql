-- account_securities.sql

-- name: CreateAccountSecurity :one
INSERT INTO account_securities (account_id)
VALUES ($1)
RETURNING *;

-- name: GetAccountSecurityByAccountID :one
SELECT * FROM account_securities WHERE account_id = $1;

-- name: GetAccountSecurity :one
SELECT * FROM account_securities WHERE id = $1;

-- name: UpdateAccountSecurity :one
UPDATE account_securities
SET account_id = $2, is_email_verified = $3, email_verified_at = $4, two_factor_enabled = $5, two_factor_enabled_at = $6
WHERE id = $1
RETURNING *;

-- name: DeleteAccountSecurity :exec
DELETE FROM account_securities WHERE id = $1;
