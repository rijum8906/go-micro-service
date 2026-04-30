--

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


-- NOTE: get methods must use 'deleted_at IS NULL'

-- name: GetOrganization :one
SELECT * FROM organizations
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetOrganizationsByCreatedBy :many
SELECT * FROM organizations
WHERE created_by = $1 AND deleted_at IS NULL
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations
WHERE slug = $1 AND deleted_at IS NULL
LIMIT 1;



-- NOTE: exists check methods must not use 'deleted_at IS NULL'

-- name: CheckOrganizationExistsBySlug :one
SELECT EXISTS(
    SELECT 1 FROM organizations WHERE slug = $1
) AS exists;

-- name: CheckOrganizationExists :one
SELECT EXISTS(
    SELECT 1 FROM organizations WHERE id = $1
) AS exists;


-- NOTE: update methods must use 'deleted_at IS NULL'

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, description = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ChangeOrganizationOwnership :exec
UPDATE organizations
SET created_by = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: ArchiveOrganization :exec
UPDATE organizations
SET status = 'archived', archived_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteOrganization :exec
UPDATE organizations
SET status = 'deleted', deleted_by = $2, deleted_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteOrganizationHard :exec
DELETE FROM organizations
WHERE id = $1;

-- name: GetDeletedOrganization :one
SELECT * FROM organizations
WHERE id = $1 AND deleted_at IS NOT NULL
LIMIT 1;
