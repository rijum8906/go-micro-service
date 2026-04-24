-- name: CreateOrganizationMembership :one
INSERT INTO organization_memberships (organization_id, user_id, role, status, invited_by, joined_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING *;

-- name: GetOrganizationMembership :one
SELECT * FROM organization_memberships
WHERE id = $1;

-- name: GetOrganizationMembershipsByUserID :many
SELECT * FROM organization_memberships
WHERE user_id = $1;

-- name: UpdateOrganizationMembership :one
UPDATE organization_memberships
SET role = $2, status = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- NOTE: soft delete
-- name: DeleteOrganizationMembership :exec
UPDATE organization_memberships
SET deleted_at = now(), deleted_by = $2
WHERE id = $1
RETURNING *;

-- name: DeleteOrganizationMembershipHard :exec
DELETE FROM organization_memberships
WHERE id = $1;
