-- name: CreateOrganizationInvitation :one
INSERT INTO organization_invitations (
    organization_id,
    email,
    role,
    invited_by_mem_id,
    token_hash,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetOrganizationInvitationByTokenHash :one
SELECT * FROM organization_invitations
WHERE token_hash = $1
  AND expires_at > NOW()
  AND status = 'pending'
LIMIT 1;

-- name: GetOrganizationInvitation :one
SELECT * FROM organization_invitations
WHERE id = $1
  AND status = 'pending'
LIMIT 1;

-- name: GetOrganizationInvitationWithAllStatus :one
SELECT * FROM organization_invitations
WHERE id = $1
LIMIT 1;

-- name: GetOrganizationInvitationsByOrgIDAndStatus :many
SELECT * FROM organization_invitations
WHERE organization_id = $1
  AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrganizationInvitationsByOrgID :many
SELECT * FROM organization_invitations
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetPendingInvitationsByEmail :many
SELECT * FROM organization_invitations
WHERE email = $1
  AND status = 'pending'
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: GetPendingInvitationByEmailAndOrg :one
SELECT * FROM organization_invitations
WHERE email = $1
  AND organization_id = $2
  AND status = 'pending'
  AND expires_at > NOW()
LIMIT 1;

-- name: CheckOrganizationInvitationExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE id = $1
) AS exists;

-- name: CheckOrganizationInvitationExistsByTokenHash :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE token_hash = $1
      AND expires_at > NOW()
      AND status = 'pending'
) AS exists;

-- name: CheckPendingInvitationExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE email = $1
      AND organization_id = $2
      AND status = 'pending'
      AND expires_at > NOW()
) AS exists;

-- name: AcceptOrganizationInvitation :one
UPDATE organization_invitations
SET
    status = 'accepted',
    responded_by_user_id = sqlc.arg(responded_by),
    responded_at = NOW(),
    response = 'accept'
WHERE id = sqlc.arg(id)
  AND status = 'pending'
  AND expires_at > NOW()
RETURNING *;

-- name: DeclineOrganizationInvitation :one
UPDATE organization_invitations
SET
    status = 'declined',
    responded_by_user_id = sqlc.arg(responded_by),
    responded_at = NOW(),
    response = 'decline'
WHERE id = sqlc.arg(id)
  AND status = 'pending'
  AND expires_at > NOW()
RETURNING *;

-- name: RevokeOrganizationInvitation :one
UPDATE organization_invitations
SET
    status = 'revoked',
    revoked_by_mem_id = $2,
    revoked_at = NOW()
WHERE id = $1
  AND status = 'pending'
  AND expires_at > NOW()
RETURNING *;

-- name: ExpireOldInvitations :exec
UPDATE organization_invitations
SET status = 'expired'
WHERE status = 'pending'
  AND expires_at <= NOW();

-- name: DeleteOrganizationInvitation :exec
DELETE FROM organization_invitations
WHERE id = $1;

-- name: DeleteExpiredInvitations :exec
DELETE FROM organization_invitations
WHERE status IN ('expired', 'accepted', 'declined', 'revoked')
  AND (
    expires_at < NOW() - INTERVAL '90 days'
    OR responded_at < NOW() - INTERVAL '90 days'
    OR revoked_at < NOW() - INTERVAL '90 days'
  );
