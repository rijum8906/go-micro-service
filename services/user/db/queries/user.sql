-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1 LIMIT 1;


-- name: UpdateUser :one
UPDATE users
SET email = $2,
password_hash = $3,
is_email_verified = $4,
email_verified_at = $5,
two_factor_enabled = $6,
two_factor_enabled_at = $7
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
