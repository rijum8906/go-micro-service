-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1 LIMIT 1;


-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserTwoFactor :one
UPDATE users
SET two_factor_enabled = $2,
    two_factor_enabled_at = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserIsEmailVerified :one
UPDATE users
SET is_email_verified = $2,
    email_verified_at = $3
WHERE id = $1
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    two_factor_enabled_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
