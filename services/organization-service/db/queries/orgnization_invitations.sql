-- name: CreateOrganizationInvitation :one
INSERT INTO organization_invitations (
    organization_id, 
    email, 
    role, 
    invited_by_mem_id, 
    token_hash, 
    expires_at,
    status  -- Added default status
) VALUES ( $1, $2, $3, $4, $5, $6, 'pending')
RETURNING *;

-- ===========================================
-- GET METHODS
-- ===========================================

-- name: GetOrganizationInvitationByTokenHash :one
SELECT * FROM organization_invitations
WHERE token_hash = $1 AND expires_at > NOW() AND status = 'pending';

-- name: GetOrganizationInvitation :one
SELECT * FROM organization_invitations
WHERE id = $1;

-- name: GetOrganizationInvitationWithAllStatus :one
SELECT * FROM organization_invitations
WHERE id = $1;

-- name: GetOrganizationInvitationsByOrgIDAndStatus :many
SELECT * FROM organization_invitations
WHERE organization_id = $1 AND status = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- name: GetOrganizationInvitationsByOrgID :many
SELECT * FROM organization_invitations
WHERE organization_id = $1
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetPendingInvitationsByEmail :many
-- Get all pending invitations for a user by email
SELECT * FROM organization_invitations
WHERE email = $1 
    AND status = 'pending' 
    AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: GetPendingInvitationByEmailAndOrg :one
-- Check for existing pending invitation (idempotency)
SELECT * FROM organization_invitations
WHERE email = $1 
    AND organization_id = $2 
    AND status = 'pending' 
    AND expires_at > NOW()
LIMIT 1;

-- ===========================================
-- CHECK METHODS
-- ===========================================

-- name: CheckOrganizationInvitationExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE id = $1
);

-- name: CheckOrganizationInvitationExistsByTokenHash :one
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE token_hash = $1 AND expires_at > NOW() AND status = 'pending'
);

-- name: CheckPendingInvitationExists :one
-- Check if there's a valid pending invitation for email+org
SELECT EXISTS (
    SELECT 1 FROM organization_invitations
    WHERE email = $1 
        AND organization_id = $2 
        AND status = 'pending' 
        AND expires_at > NOW()
);

-- ===========================================
-- UPDATE METHODS
-- ===========================================

-- name: AcceptOrganizationInvitation :one
UPDATE organization_invitations
SET 
    status = 'accepted', 
    responded_by = $2, 
    responded_at = NOW(), 
    response = 'accept'
WHERE id = $1 
    AND status = 'pending'  -- Only accept pending invitations
    AND expires_at > NOW()   -- Only accept if not expired
RETURNING *;

-- name: DeclineOrganizationInvitation :one
UPDATE organization_invitations
SET 
    status = 'declined', 
    responded_by = $2, 
    responded_at = NOW(), 
    response = 'decline'
WHERE id = $1 
    AND status = 'pending'  -- Only decline pending invitations
    AND expires_at > NOW()   -- Only decline if not expired
RETURNING *;

-- name: RevokeOrganizationInvitation :one
UPDATE organization_invitations
SET 
    status = 'revoked', 
    revoked_by_mem_id = $2, 
    revoked_at = NOW()
WHERE id = $1 
    AND status = 'pending'  -- Only revoke pending invitations
    AND expires_at > NOW()   -- Only revoke if not expired
RETURNING *;

-- name: ExpireOldInvitations :exec
-- Clean up expired invitations (can be run as a background job)
UPDATE organization_invitations
SET status = 'expired'
WHERE status = 'pending' AND expires_at <= NOW();

-- WARNING: Hard delete - use with caution
-- name: DeleteOrganizationInvitation :exec
DELETE FROM organization_invitations
WHERE id = $1;

-- name: DeleteExpiredInvitations :exec
-- Clean up expired invitations permanently (background job)
DELETE FROM organization_invitations
WHERE status IN ('expired', 'accepted', 'declined', 'revoked')
    AND (expires_at < NOW() - INTERVAL '90 days' OR responded_at < NOW() - INTERVAL '90 days');
