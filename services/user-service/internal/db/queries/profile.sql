-- profile.sql

-- name: CreateProfile :one
INSERT INTO profiles (account_id, first_name, last_name, display_name, avatar_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProfileByAccountID :one
SELECT * FROM profiles WHERE account_id = $1;

-- name: GetProfile :one
SELECT * FROM profiles WHERE id = $1;

-- name: GetProfilesByAccountID :many
SELECT * FROM profiles WHERE account_id = $1;

-- name: UpdateProfile :one
UPDATE profiles
SET 
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    display_name = COALESCE(sqlc.narg('display_name'), display_name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE account_id = $1;
