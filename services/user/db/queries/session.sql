-- name: GetSession :one
SELECT *
FROM sessions
WHERE id = $1 LIMIT 1;

-- name: GetSessionsByUserID :many
SELECT *
FROM sessions
WHERE user_id = $1 ORDER BY last_login_at DESC LIMIT $2 OFFSET $3;

-- name: GetActiveSessionsByUserID :many
SELECT *
FROM sessions
WHERE user_id = $1 AND is_revoked = false ORDER BY last_login_at DESC LIMIT $2 OFFSET $3;

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

-- name: RevokeSession :exec
UPDATE sessions
SET is_revoked = true
WHERE id = $1;

-- name: RevokeActiveSessions :exec
UPDATE sessions
SET is_revoked = true
WHERE user_id = $1 AND is_revoked = false;

-- name: RevokeOtherSessions :exec
UPDATE sessions
SET is_revoked = true
WHERE user_id = $1 AND id <> $2 AND is_revoked = false;

-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    refresh_token_hash,
    user_agent,
    ip_addr,
    device_id,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= now();

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;
