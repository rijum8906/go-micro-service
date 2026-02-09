-- session.sql

-- name: CreateSession :one
INSERT INTO sessions (account_id, refresh_token, user_agent, ip_addr, device_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionByRefreshToken :one
SELECT * FROM sessions WHERE refresh_token = $1 AND is_revoked = FALSE;

-- name: GetSessionByAccountID :one
SELECT * FROM sessions WHERE account_id = $1 ORDER BY last_login_at DESC LIMIT 1;

-- name: GetSessionsByAccountID :many
SELECT * FROM sessions WHERE account_id = $1 ORDER BY last_login_at DESC LIMIT $2 OFFSET $3;

-- name: RevokeSession :exec
UPDATE sessions 
SET is_revoked = TRUE, updated_at = NOW() 
WHERE id = $1;

-- name: RevokeAllAccountSessions :exec
UPDATE sessions 
SET is_revoked = TRUE, updated_at = NOW() 
WHERE account_id = $1 AND is_revoked = FALSE;

-- name: UpdateSession :one
UPDATE sessions
SET 
    refresh_token = COALESCE(sqlc.narg('refresh_token'), refresh_token),
    user_agent = COALESCE(sqlc.narg('user_agent'), user_agent),
    ip_addr = COALESCE(sqlc.narg('ip_addr'), ip_addr),
    device_id = COALESCE(sqlc.narg('device_id'), device_id),
    last_login_at = COALESCE(sqlc.narg('last_login_at'), last_login_at),
    is_revoked = COALESCE(sqlc.narg('is_revoked'), is_revoked),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;
