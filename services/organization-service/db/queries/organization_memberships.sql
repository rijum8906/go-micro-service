-- name: CreateOrganizationMembership :one
INSERT INTO organization_memberships (
    user_id, organization_id, role
) VALUES ( $1, $2, $3 )
    RETURNING *;

-- name: GetOrganizationMembershipsByUserID :many
SELECT * FROM organization_memberships
WHERE user_id = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CreateOrganizationMembershipOwner :one
INSERT INTO organization_memberships (
    user_id, organization_id, role
) VALUES ( $1, $2, $3 )
    RETURNING *;

-- name: DeleteOrganizationMembership :exec
UPDATE organization_memberships
SET status = 'deleted', deleted_by = $2, deleted_at = now()
WHERE id = $1;

-- name: DeleteOrganizationMembershipHard :exec
DELETE FROM organization_memberships
WHERE id = $1;

