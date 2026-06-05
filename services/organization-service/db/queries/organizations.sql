-- name: CreateOrganization :one
INSERT INTO organizations (
    name,
    slug,
    description,
    created_by_user_id
) VALUES (
    $1, $2, $3, sqlc.arg(created_by)
)
RETURNING *;

-- name: GetOrganization :one
SELECT * FROM organizations
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetOrganizationsByCreatedBy :many
SELECT * FROM organizations
WHERE created_by_user_id = sqlc.arg(created_by)
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations
WHERE slug = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetDeletedOrganization :one
SELECT * FROM organizations
WHERE id = $1
  AND deleted_at IS NOT NULL
LIMIT 1;

-- name: CheckOrganizationExistsBySlug :one
SELECT EXISTS (
    SELECT 1 FROM organizations
    WHERE slug = $1
) AS exists;

-- name: CheckOrganizationExists :one
SELECT EXISTS (
    SELECT 1 FROM organizations
    WHERE id = $1
      AND deleted_at IS NULL
) AS exists;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = $2,
    description = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ChangeOrganizationOwnership :exec
UPDATE organizations
SET
    created_by_user_id = sqlc.arg(created_by),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ArchiveOrganization :exec
UPDATE organizations
SET
    status = 'archived',
    archived_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'active';

-- name: DeleteOrganization :exec
UPDATE organizations
SET
    status = 'deleted',
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: RestoreArchivedOrganization :exec
UPDATE organizations
SET
    status = 'active',
    archived_at = NULL,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'archived';

-- name: DeleteOrganizationHard :exec
DELETE FROM organizations
WHERE id = $1;
