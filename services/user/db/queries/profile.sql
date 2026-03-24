-- name: GetProfile :one
SELECT *
FROM profiles
WHERE user_id = $1 LIMIT 1;

-- name: GetProfileUserID :one
SELECT *
FROM profiles
WHERE user_id = $1 LIMIT 1;

-- name: UpdateProfile :one
UPDATE profiles
SET first_name = $2,
last_name = $3,
avatar_url = $4
WHERE id = $1
RETURNING *;

-- name: CreateProfile :one
INSERT INTO profiles (
    user_id,
    first_name,
    last_name,
    avatar_url
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteProfile :exec
DELETE FROM profiles
WHERE id = $1;
