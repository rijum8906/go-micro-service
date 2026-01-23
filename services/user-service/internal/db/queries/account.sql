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
SET email = $2, password_hash = $3
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;
