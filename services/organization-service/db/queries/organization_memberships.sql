-- name: CreateOrganizationMembership :one
INSERT INTO organization_memberships (
    user_id, organization_id, role
) VALUES ( $1, $2, $3 )
    RETURNING *;

-- name: CreateOrganizationMembershipOwner :one
INSERT INTO organization_memberships (
    user_id, organization_id, role
) VALUES ( $1, $2, $3 )
    RETURNING *;


-- NOTE: get and update methods must use "status != 'deleted'"

-- name: GetOrganizationMembership :one
SELECT * FROM organization_memberships
WHERE id = $1 AND status != 'deleted';

-- name: GetOrganizationMembershipsByUserID :many
SELECT * FROM organization_memberships
WHERE user_id = $1 AND status != 'deleted'
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1 AND status != 'deleted'
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipsByOrgIDAndRole :many
SELECT * FROM organization_memberships
WHERE organization_id = $1 AND role = $2 AND status != 'deleted'
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- NOTE: this method also return deleted so do not use "status != 'deleted'"
-- name: GetOrganizationMembershipsByOrgIDAndStatus :many
SELECT * FROM organization_memberships
WHERE organization_id = $1 AND status = $2
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: UpdateOrganizationMembershipRole :one
UPDATE organization_memberships
SET role = $2
WHERE id = $1 AND status != 'deleted'
RETURNING *;

-- name: UpdateOrganizationMembershipStatus :one
UPDATE organization_memberships
SET status = $2
WHERE id = $1 AND status != 'deleted'
RETURNING *;

-- name: DeleteOrganizationMembership :exec
UPDATE organization_memberships
SET status = 'deleted', deleted_by = $2, deleted_at = now()
WHERE id = $1;

-- name: DeleteOrganizationMembershipHard :exec
DELETE FROM organization_memberships
WHERE id = $1;
