-- name: GetSession :one
SELECT *
FROM sessions
WHERE id = $1 LIMIT 1;

-- name: GetSessionUserID :one
SELECT *
FROM sessions
WHERE user_id = $1 LIMIT 1;

-- name: GetSessionByRefreshTokenHash :one
SELECT *
FROM sessions
WHERE refresh_token_hash = $1 LIMIT 1;

-- name: UpdateSession :one
UPDATE sessions
SET user_id = $2,
refresh_token_hash = $3,
user_agent = $4,
ip_addr = $5,
device_id = $6,
last_login_at = $7,
expires_at = $8,
is_revoked = $9
WHERE id = $1
RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    refresh_token_hash,
    user_agent,
    ip_addr,
    device_id,
    last_login_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;
