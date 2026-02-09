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
SET 
  is_email_verified = COALESCE(sqlc.narg('is_email_verified'), is_email_verified),
  email_verified_at = COALESCE(sqlc.narg('email_verified_at'), email_verified_at),
  two_factor_enabled = COALESCE(sqlc.narg('two_factor_enabled'), two_factor_enabled),
  two_factor_enabled_at = COALESCE(sqlc.narg('two_factor_enabled_at'), two_factor_enabled_at),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateAccountSecurityByAccountID :one
UPDATE account_securities
SET 
  is_email_verified = COALESCE(sqlc.narg('is_email_verified'), is_email_verified),
  email_verified_at = COALESCE(sqlc.narg('email_verified_at'), email_verified_at),
  two_factor_enabled = COALESCE(sqlc.narg('two_factor_enabled'), two_factor_enabled),
  two_factor_enabled_at = COALESCE(sqlc.narg('two_factor_enabled_at'), two_factor_enabled_at),
  updated_at = NOW()
WHERE account_id = $1
RETURNING *;

-- name: DeleteAccountSecurity :exec
DELETE FROM account_securities WHERE id = $1;
