-- account.sql

-- name: CreateAccount :one
INSERT INTO accounts (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetAccountByEmail :one
SELECT * FROM accounts WHERE email = $1;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1;
