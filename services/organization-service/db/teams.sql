-- name: CreateOrganizationTeam :one
INSERT INTO organization_teams (organization_id, name, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrganizationTeam :one
SELECT * FROM organization_teams
WHERE id = $1;

-- name: GetOrganizationTeamsByOrganizationID :many
SELECT * FROM organization_teams
WHERE organization_id = $1;

-- name: UpdateOrganizationTeam :one
UPDATE organization_teams
SET name = $2, description = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- NOTE: soft delete
-- name: DeleteOrganizationTeam :exec
UPDATE organization_teams
SET deleted_at = now(), deleted_by = $2
WHERE id = $1
RETURNING *;

-- name: DeleteOrganizationTeamHard :exec
DELETE FROM organization_teams
WHERE id = $1;
