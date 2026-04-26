-- name: CreateOrganizationMembership :one
INSERT INTO organization_memberships (
    user_id, organization_id, role, invited_by
) VALUES ( $1, $2, $3, $4 )
    RETURNING *;

-- name: CreateOrganizationMembershipOwner :one
INSERT INTO organization_memberships (
    user_id, organization_id, role
) VALUES ( $1, $2, $3 )
    RETURNING *;

