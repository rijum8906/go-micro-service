-- ===========================================
-- CREATE METHODS
-- ===========================================
-- name: CreateTwoFactorAuth :one
INSERT INTO two_factors
(user_id, method, secret)
VALUES ($1, $2, $3)
RETURNING *;

-- ===========================================
-- GET METHODS
-- ===========================================
-- name: GetTwoFactorAuth :one
SELECT * FROM two_factors WHERE id = $1;

-- name: GetTwoFactorAuthByUserID :one
SELECT * FROM two_factors WHERE user_id = $1;

-- name: GetPrimaryTwoFactorAuthByUserID :one
SELECT * FROM two_factors WHERE user_id = $1 AND is_primary = true;

-- ===========================================
-- CHECK METHODS
-- ===========================================
-- name: CheckTwoFactorAuthEnabledByUserID :one
SELECT EXISTS(
    SELECT 1 FROM two_factors WHERE user_id = $1 AND is_enabled = true
);

-- name: CheckTwoFactorAuthEnabledByUserIDAndMethod :one
SELECT EXISTS(
    SELECT 1 FROM two_factors WHERE user_id = $1 AND method = $2 AND is_enabled = true
);

-- name: CheckTwoFactorAuthEnabledBySecret :one
SELECT EXISTS(
    SELECT 1 FROM two_factors WHERE user_id = $1 AND secret = $2 AND is_enabled = true
);

-- ============================================
-- UPDATE METHODS
-- ============================================
-- name: SetPrimaryTwoFactorSecretByUserID :exec
UPDATE two_factors
SET secret = $2, is_primary = true
WHERE user_id = $1;

-- name: SetSecondaryTwoFactorSecretByUserID :exec
UPDATE two_factors
SET secret = $2, is_primary = false
WHERE user_id = $1;

-- name: EnableTwoFactorAuthByUserIDAndMethod :exec
UPDATE two_factors
SET is_enabled = true
WHERE user_id = $1 AND method = $2;


-- name: DisableTwoFactorAuthByUserIDAndMethod :exec
UPDATE two_factors
SET is_enabled = false
WHERE user_id = $1 AND method = $2;

-- name: EnableTwoFactorAuthsByUserID :exec
UPDATE two_factors
SET is_enabled = true
WHERE user_id = $1;

-- name: DisableTwoFactorAuthsByUserID :exec
UPDATE two_factors
SET is_enabled = false
WHERE user_id = $1;

-- ============================================
-- DELETE METHODS
-- ============================================
-- name: DeleteTwoFactorAuthHardByUserID :exec
DELETE FROM two_factors WHERE user_id = $1;
