-- account.sql

-- name: CreateAccount :one
INSERT INTO accounts (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetAccountByEmail :one
SELECT * FROM accounts WHERE email = $1;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1;

-- name: UpdateAccount :one
UPDATE accounts
SET 
  email = COALESCE(sqlc.narg('email'), email),
  password_hash = COALESCE(sqlc.narg('passwordHash'), password_hash),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateAccountByEmail :one
UPDATE accounts
SET 
  email = COALESCE(sqlc.narg('new_email'), email),
  password_hash = COALESCE(sqlc.narg('passwordHash'), password_hash),
  updated_at = NOW()
WHERE email = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;
