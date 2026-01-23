-- oauth.sql
-- name: CreateOAuth :one
INSERT INTO oauths(
  account_id,
  provider,
  subject,
  token
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOAuthBySubject :one
SELECT * FROM oauths WHERE subject = $1;

-- name: GetOAuthsByAccountID :one
SELECT * FROM oauths WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: GetOAuthBySubjectAndProvider :one
SELECT * FROM oauths WHERE subject = $1 AND provider = $2;

-- name: GetOAuth :one
SELECT * FROM oauths WHERE id = $1;

-- name: GetOAuthByAccountID :one
SELECT * FROM oauths WHERE account_id = $1;

-- name: UpdateOAuth :one
UPDATE oauths
SET account_id = $2, provider = $3, subject = $4, token = $5 
WHERE id = $1
RETURNING *;

-- name: DeleteOAuth :exec
DELETE FROM oauths WHERE id = $1;
