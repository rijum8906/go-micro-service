-- =====================================================
-- CREATE METHODS
-- =====================================================
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

-- =====================================================
-- GET METHODS
-- =====================================================
-- name: GetSession :one
SELECT *
    FROM sessions
    WHERE id = $1 AND is_revoked = false LIMIT 1;

-- name: GetSessionsByUserID :many
SELECT *
FROM sessions
WHERE user_id = $1 AND is_revoked = false ORDER BY last_login_at DESC LIMIT $2 OFFSET $3;

-- name: GetActiveSessionsByUserID :many
SELECT *
FROM sessions
WHERE user_id = $1 AND is_revoked = false ORDER BY last_login_at DESC LIMIT $2 OFFSET $3;

-- name: GetSessionByRefreshTokenHash :one
SELECT *
FROM sessions
WHERE refresh_token_hash = $1 AND is_revoked = false LIMIT 1;

-- =====================================================
-- CHECK METHODS
-- =====================================================
-- name: CheckSessionExists :one
SELECT EXISTS(
    SELECT 1
    FROM sessions
    WHERE id = $1 AND is_revoked = false
);

-- name: CheckSessionExistsByTokenHash :one
SELECT EXISTS(
    SELECT 1
    FROM sessions
    WHERE refresh_token_hash = $1 AND is_revoked = false
);

-- =====================================================
-- UPDATE METHODS
-- =====================================================
-- name: UpdateSessionRefreshTokenHash :one
UPDATE sessions
SET refresh_token_hash = $2,
last_login_at = NOW()
WHERE id = $1 AND is_revoked = false
RETURNING *;

-- name: RevokeSession :exec
UPDATE sessions
SET is_revoked = true
WHERE id = $1 AND is_revoked = false;

-- name: RevokeSessionByRefreshTokenHash :exec
UPDATE sessions
SET is_revoked = true
WHERE refresh_token_hash = $1 AND is_revoked = false;

-- name: RevokeActiveSessions :exec
UPDATE sessions
SET is_revoked = true
WHERE user_id = $1 AND is_revoked = false;

-- name: RevokeOtherSessions :exec
UPDATE sessions
SET is_revoked = true
WHERE user_id = $1 AND id = $2 AND is_revoked = false;

-- =====================================================
-- LOCK METHODS
-- =====================================================
-- name: LockSession :exec
SELECT *
FROM sessions
WHERE id = $1 AND is_revoked = false
FOR UPDATE SKIP LOCKED;

-- name: LockAndGetSessionByRefreshTokenHash :one
SELECT *
FROM sessions
WHERE refresh_token_hash = $1 AND is_revoked = false
FOR UPDATE SKIP LOCKED;

-- name: LockSessionsByUserID :exec
SELECT *
FROM sessions
WHERE user_id = $1 AND is_revoked = false
FOR UPDATE SKIP LOCKED;

-- =====================================================
-- DELETE METHODS
-- =====================================================
-- name: DeleteSessionHard :exec
DELETE FROM sessions
WHERE id = $1 AND is_revoked = false;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= now() AND is_revoked = false;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1 AND is_revoked = false;
