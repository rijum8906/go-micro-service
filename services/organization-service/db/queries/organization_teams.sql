-- name: CreateOrganizationTeam :one
INSERT INTO organization_teams (
    organization_id,
    name,
    description,
    created_by_mem_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetOrganizationTeam :one
SELECT * FROM organization_teams
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetOrganizationTeamByOrgIDAndName :one
SELECT * FROM organization_teams
WHERE organization_id = $1
  AND name = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetDeletedOrganizationTeam :one
SELECT * FROM organization_teams
WHERE id = $1
  AND deleted_at IS NOT NULL
LIMIT 1;

-- name: CheckOrganizationTeamExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_teams
    WHERE id = $1
      AND deleted_at IS NULL
) AS exists;

-- name: CheckOrganizationTeamNameExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_teams
    WHERE organization_id = $1
      AND name = $2
      AND status IN ('active', 'archived')
) AS exists;

-- name: GetOrganizationTeamsByOrgID :many
SELECT * FROM organization_teams
WHERE organization_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationTeamsByOrgIDAndStatus :many
SELECT * FROM organization_teams
WHERE organization_id = $1
  AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrganizationTeamsByCreatedBy :many
SELECT * FROM organization_teams
WHERE created_by_mem_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrganizationTeamsByOrgID :one
SELECT COUNT(*) FROM organization_teams
WHERE organization_id = $1
  AND status = 'active';

-- name: CountOrganizationTeamsByOrgIDAndStatus :one
SELECT COUNT(*) FROM organization_teams
WHERE organization_id = $1
  AND status = $2;

-- name: UpdateOrganizationTeam :one
UPDATE organization_teams
SET
    name = $2,
    description = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveOrganizationTeam :exec
UPDATE organization_teams
SET
    status = 'archived',
    archived_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'active';

-- name: RestoreArchivedOrganizationTeam :exec
UPDATE organization_teams
SET
    status = 'active',
    archived_at = NULL,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'archived';

-- name: DeleteOrganizationTeam :exec
UPDATE organization_teams
SET
    status = 'deleted',
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: DeleteOrganizationTeamHard :exec
DELETE FROM organization_teams
WHERE id = $1;

-- name: DeleteOrganizationTeamsByOrgIDHard :exec
DELETE FROM organization_teams
WHERE organization_id = $1;
