-- session.sql

-- name: CreateSession :one
INSERT INTO sessions (account_id, refresh_token, user_agent, ip_addr, device_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionByRefreshToken :one
SELECT * FROM sessions WHERE refresh_token = $1;

-- name: GetSessionByAccountID :one
SELECT * FROM sessions WHERE account_id = $1 ORDER BY last_login_at DESC LIMIT 1;

-- name: GetSessionsByAccountID :many
SELECT * FROM sessions WHERE account_id = $1 ORDER BY last_login_at DESC LIMIT $2 OFFSET $3;

-- name: RevokeSession :exec
UPDATE sessions SET is_revoked = TRUE WHERE id = $1;

-- name: UpdateSession :one
UPDATE sessions
SET account_id = $2, refresh_token = $3, user_agent = $4, ip_addr = $5, device_id = $6, last_login_at = $7
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;
