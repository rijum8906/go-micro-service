-- name: CreateOrganizationInvitation :one
INSERT INTO organization_invitations (organization_id, email, role, invited_by, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOrganizationInvitationsByOrganizationID :one
SELECT *
FROM organization_invitations
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationInvitationsByEmail :one
SELECT *
FROM organization_invitations
WHERE email = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationInvitation :one
SELECT *
FROM organization_invitations
WHERE id = $1;

-- name: AcceptOrganizationInvitation :one
UPDATE organization_invitations
SET status = 'accepted', responded_by = $2, accepted_at = now()
WHERE id = $1
RETURNING *;



-- name: DeleteOrganizationInvitation :exec
DELETE FROM organization_invitations
WHERE id = $1;
