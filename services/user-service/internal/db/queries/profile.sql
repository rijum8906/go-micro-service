-- profile.sql

-- name: CreateProfile :one
INSERT INTO profiles (account_id, first_name, last_name, display_name, avatar_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProfile :one
SELECT * FROM profiles WHERE account_id = $1;

-- name: GetProfileByAccountID :one
SELECT * FROM profiles WHERE account_id = $1;

-- name: GetProfilesByAccountID :many
SELECT * FROM profiles WHERE account_id = $1;

-- name: UpdateProfile :one
UPDATE profiles
SET first_name = $2, last_name = $3, display_name = $4, avatar_url = $5
WHERE id = $1
RETURNING *;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE account_id = $1;
