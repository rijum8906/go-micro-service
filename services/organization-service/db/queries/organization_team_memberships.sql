-- name: CreateOrganizationTeamMembership :one
INSERT INTO organization_team_memberships (
    organization_id,
    team_id,
    membership_id,
    role
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: CheckOrganizationTeamMembershipExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_team_memberships
    WHERE id = $1
      AND status NOT IN ('left', 'removed')
) AS exists;

-- name: GetOrganizationTeamMembership :one
SELECT * FROM organization_team_memberships
WHERE id = $1
  AND status NOT IN ('left', 'removed')
LIMIT 1;

-- name: GetOrganizationTeamMembershipByTeamIDAndMembershipID :one
SELECT * FROM organization_team_memberships
WHERE team_id = $1
  AND membership_id = $2
  AND status NOT IN ('left', 'removed')
LIMIT 1;

-- name: GetOrganizationTeamMembershipsByTeamID :many
SELECT * FROM organization_team_memberships
WHERE team_id = $1
  AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationTeamMembershipsByMembershipID :many
SELECT * FROM organization_team_memberships
WHERE membership_id = $1
  AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationTeamMembershipWithAllStatuses :one
SELECT * FROM organization_team_memberships
WHERE id = $1
LIMIT 1;

-- name: CountActiveTeamMembersByTeamID :one
SELECT COUNT(*) FROM organization_team_memberships
WHERE team_id = $1
  AND status = 'active';

-- name: UpdateOrganizationTeamMembershipRole :one
UPDATE organization_team_memberships
SET
    role = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING *;

-- name: UpdateOrganizationTeamMembershipStatus :one
UPDATE organization_team_memberships
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
  AND status NOT IN ('left', 'removed')
RETURNING *;

-- name: UpdateOrganizationTeamMembershipStatusWithAllStatuses :one
UPDATE organization_team_memberships
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: LockOrganizationTeamMembershipForUpdate :one
SELECT id FROM organization_team_memberships
WHERE id = $1
FOR UPDATE;

-- name: SoftDeleteOrganizationTeamMembership :exec
UPDATE organization_team_memberships
SET
    status = 'left',
    updated_at = NOW()
WHERE id = $1
  AND status NOT IN ('left', 'removed');

-- name: SoftDeleteOrganizationTeamMembershipByTeamIDAndMembershipID :exec
UPDATE organization_team_memberships
SET
    status = 'left',
    updated_at = NOW()
WHERE team_id = $1
  AND membership_id = $2
  AND status NOT IN ('left', 'removed');

-- name: HardDeleteOrganizationTeamMembership :exec
DELETE FROM organization_team_memberships
WHERE id = $1;

-- name: HardDeleteOrganizationTeamMembershipsByTeamID :exec
DELETE FROM organization_team_memberships
WHERE team_id = $1;

-- name: GetActiveTeamMembershipsByTeamID :many
SELECT * FROM organization_team_memberships
WHERE team_id = $1
  AND status = 'active'
ORDER BY created_at DESC;

-- name: GetTeamMembershipsByMembershipIDs :many
SELECT * FROM organization_team_memberships
WHERE membership_id = ANY($1::uuid[])
  AND status NOT IN ('left', 'removed');

-- name: GetTeamMembershipsByTeamIDs :many
SELECT * FROM organization_team_memberships
WHERE team_id = ANY($1::uuid[])
  AND status NOT IN ('left', 'removed');

-- name: GetAllOrganizationTeamMembershipsByTeamID :many
SELECT * FROM organization_team_memberships
WHERE team_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
