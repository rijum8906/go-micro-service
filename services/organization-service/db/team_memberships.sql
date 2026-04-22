-- name: CreateOrganizationTeamMember :one
INSERT INTO organization_team_memberships (team_id, membership_id, role, invited_by, joined_at)
VALUES ($1, $2, $3, $4, now())
RETURNING *;

-- GetOrganizationTeamMembershipsByTeamID :many
SELECT * FROM organization_team_memberships
WHERE team_id = $1;

-- name: GetOrganizationTeamMembership :one
SELECT * FROM organization_team_memberships
WHERE id = $1;

-- name: GetOrganizationTeamMembershipsByMembershipID :many
SELECT * FROM organization_team_memberships
WHERE membership_id = $1;

-- name: UpdateOrganizationTeamMembership :one
UPDATE organization_team_memberships
SET role = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- NOTE: soft delete
-- name: DeleteOrganizationTeamMembership :exec
UPDATE organization_team_memberships
SET deleted_at = now(), deleted_by = $2
WHERE id = $1
RETURNING *;

-- name: DeleteOrganizationTeamMembershipHard :exec
DELETE FROM organization_team_memberships
WHERE id = $1;
