-- name: CreateOrganizationInvitation :one
INSERT INTO organization_invitations (
    organization_id, email, role, invited_by, token_hash, expires_at
) VALUES ( $1, $2, $3, $4, $5, $6)
    RETURNING *;

-- name: GetOrganizationInvitationByTokenHash :one
SELECT * FROM organization_invitations
WHERE token_hash = $1;

-- name: GetOrganizationInvitation :one
SELECT * FROM organization_invitations
WHERE id = $1;

-- name: GetOrganizationInvitationsByOrgIDAndStatus :many
SELECT * FROM organization_invitations
WHERE organization_id = $1 AND status = $2
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: GetOrganizationInvitationsByOrgID :many
SELECT * FROM organization_invitations
WHERE organization_id = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CheckOrganizationInvitationExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE id = $1
);

-- name: CheckOrganizationInvitationExistsByTokenHash :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE token_hash = $1
);

-- name: AccecptOrganizationInvitation :one
UPDATE organization_invitations
SET status = 'accepted', responded_by = $2, responded_at = now(), response = 'accept'
WHERE id = $1
RETURNING *;

-- name: DeclineOrganizationInvitation :one
UPDATE organization_invitations SET status = 'declined', responded_by = $2, responded_at = now(), response = 'decline'
WHERE id = $1
RETURNING *;

-- name: RevokeOrganizationInvitation :one
UPDATE organization_invitations
SET status = 'revoked', revoked_by = $2, revoked_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteOrganizationInvitation :exec
DELETE FROM organization_invitations
WHERE id = $1;
