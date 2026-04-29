-- name: CreateOrganization :one
INSERT INTO organizations (
    name,
    slug,
    description,
    created_by
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetOrganization :one
SELECT * FROM organizations
WHERE id = $1
LIMIT 1;

-- name: GetOrganizationsByCreatedBy :many
SELECT * FROM organizations
WHERE created_by = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations
WHERE slug = $1
LIMIT 1;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteOrganization :exec
UPDATE organizations
SET status = 'deleted', deleted_by = $2, deleted_at = now()
WHERE id = $1;

-- name: DeleteOrganizationHard :exec
DELETE FROM organizations
WHERE id = $1;
