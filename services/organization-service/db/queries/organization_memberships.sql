-- name: CreateOrganizationMembership :one
INSERT INTO organization_memberships (
    user_id,
    organization_id,
    role
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: CreateOrganizationMembershipOwner :one
INSERT INTO organization_memberships (
    user_id,
    organization_id,
    role
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: CheckOrganizationMembershipExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_memberships
    WHERE id = $1
      AND status NOT IN ('left', 'removed')
) AS exists;

-- name: GetOrganizationMembership :one
SELECT * FROM organization_memberships
WHERE id = $1
  AND status NOT IN ('left', 'removed')
LIMIT 1;

-- name: GetOrganizationMembershipByOrgIDAndUserID :one
SELECT * FROM organization_memberships
WHERE organization_id = $1
  AND user_id = $2
  AND status NOT IN ('left', 'removed')
LIMIT 1;

-- name: GetOrganizationMembershipByUserIDAndOrgID :one
SELECT * FROM organization_memberships
WHERE user_id = $1
  AND organization_id = $2
  AND status NOT IN ('left', 'removed')
LIMIT 1;

-- name: GetOrganizationMembershipsByUserID :many
SELECT * FROM organization_memberships
WHERE user_id = $1
  AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
  AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipsByOrgIDAndRole :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
  AND role = $2
  AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrganizationMembershipsByOrgIDAndStatus :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
  AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrganizationMembershipWithAllStatuses :one
SELECT * FROM organization_memberships
WHERE id = $1
LIMIT 1;

-- name: CountActiveOwnersByOrgID :one
SELECT COUNT(*) FROM organization_memberships
WHERE organization_id = $1
  AND role = 'owner'
  AND status NOT IN ('left', 'removed');

-- name: CountActiveMembersByOrgID :one
SELECT COUNT(*) FROM organization_memberships
WHERE organization_id = $1
  AND status = 'active';

-- name: CountMembersByOrgIDAndStatus :one
SELECT COUNT(*) FROM organization_memberships
WHERE organization_id = $1
  AND status = $2;

-- name: UpdateOrganizationMembershipRole :one
UPDATE organization_memberships
SET
    role = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING *;

-- name: UpdateOrganizationMembershipStatus :one
UPDATE organization_memberships
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
  AND status NOT IN ('left', 'removed')
RETURNING *;

-- name: UpdateOrganizationMembershipStatusWithAllStatuses :one
UPDATE organization_memberships
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: LockOrganizationMembershipForUpdate :one
SELECT id FROM organization_memberships
WHERE id = $1
FOR UPDATE;

-- name: SoftDeleteOrganizationMembership :exec
UPDATE organization_memberships
SET
    status = 'left',
    updated_at = NOW()
WHERE id = $1
  AND status NOT IN ('left', 'removed');

-- name: SoftDeleteOrganizationMembershipByOrgIDAndUserID :exec
UPDATE organization_memberships
SET
    status = 'left',
    updated_at = NOW()
WHERE organization_id = $1
  AND user_id = $2
  AND status NOT IN ('left', 'removed');

-- name: RemoveOrganizationMembership :exec
UPDATE organization_memberships
SET
    status = 'removed',
    updated_at = NOW()
WHERE id = $1
  AND status NOT IN ('left', 'removed');

-- name: RestoreOrganizationMembership :one
UPDATE organization_memberships
SET
    status = 'active',
    updated_at = NOW()
WHERE id = $1
  AND status = 'left'
RETURNING *;

-- name: HardDeleteOrganizationMembership :exec
DELETE FROM organization_memberships
WHERE id = $1;

-- name: HardDeleteOrganizationMembershipByOrgID :exec
DELETE FROM organization_memberships
WHERE organization_id = $1;

-- name: GetActiveMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
  AND status = 'active'
ORDER BY created_at DESC;

-- name: GetBannedMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
  AND status = 'banned'
ORDER BY created_at DESC;

-- name: GetMembershipsByUserIDs :many
SELECT * FROM organization_memberships
WHERE user_id = ANY($1::uuid[])
  AND status NOT IN ('left', 'removed');

-- name: GetMembershipsByOrgIDs :many
SELECT * FROM organization_memberships
WHERE organization_id = ANY($1::uuid[])
  AND status NOT IN ('left', 'removed');

-- name: GetAllOrganizationMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipHistory :many
SELECT * FROM organization_memberships
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
