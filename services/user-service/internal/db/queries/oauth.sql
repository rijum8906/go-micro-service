-- oauth.sql

-- name: CreateOAuth :one
INSERT INTO oauths(
  account_id,
  provider,
  subject
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOAuthBySubject :one
SELECT * FROM oauths WHERE subject = $1;

-- name: GetOAuthsByAccountID :many
SELECT * FROM oauths WHERE account_id = $1;

-- name: GetOAuthBySubjectAndProvider :one
SELECT * FROM oauths WHERE subject = $1 AND provider = $2;

-- name: GetOAuth :one
SELECT * FROM oauths WHERE id = $1;

-- name: GetOAuthByAccountID :one
SELECT * FROM oauths WHERE account_id = $1;

-- name: UpdateOAuth :one
UPDATE oauths
SET 
  provider = COALESCE(sqlc.narg('provider'), provider),
  subject = COALESCE(sqlc.narg('subject'), subject),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOAuthBySubject :one
UPDATE oauths
SET 
  updated_at = NOW()
WHERE subject = $1
RETURNING *;

-- name: DeleteOAuth :exec
DELETE FROM oauths WHERE id = $1;
