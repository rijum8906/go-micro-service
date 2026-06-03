-- =====================================================
-- CREATE METHODS
-- =====================================================
-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash
) VALUES (
    $1, $2
)
RETURNING *;

-- =====================================================
-- GET METHODS
-- =====================================================
-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1 AND status != 'deleted' LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1 AND status != 'deleted' LIMIT 1;

-- name: GetUserWithAllStatuses :one
SELECT *
FROM users
WHERE id = $1;

-- =====================================================
-- CHECK METHODS
-- =====================================================
-- name: CheckUserExists :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE id = $1 AND status != 'deleted'
);

-- name: CheckUserEmailExists :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE email = $1 AND status != 'deleted'
);

-- name: CheckUserEmailExistsWithAllStatuses :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE email = $1
);

-- =====================================================
-- UPDATE METHODS
-- =====================================================
-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: UpdateUserEmail :exec
UPDATE users
SET email = $2
WHERE id = $1;

-- name: UpdateUserIsEmailVerified :exec
UPDATE users
SET is_email_verified = $2,
    email_verified_at = $3
WHERE id = $1;

-- =====================================================
-- DELETE METHODS
-- =====================================================
-- name: DeleteUserSoft :exec
UPDATE users
SET status = 'deleted', deleted_at = NOW()
WHERE id = $1;

-- name: DeleteUserHard :exec
DELETE FROM users
WHERE id = $1;
