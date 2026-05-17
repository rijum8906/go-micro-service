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

-- ============================================
-- GET METHODS (exclude left/removed by default)
-- ============================================

-- name: CheckOrganizationMembershipExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_memberships
    WHERE id = $1 AND status NOT IN ('left', 'removed')
);

-- name: GetOrganizationMembership :one
SELECT * FROM organization_memberships
WHERE id = $1 AND status NOT IN ('left', 'removed');

-- name: GetOrganizationMembershipByUserIDAndOrgID :one
SELECT * FROM organization_memberships
WHERE user_id = $1 AND organization_id = $2 AND status NOT IN ('left', 'removed');

-- name: GetOrganizationMembershipByOrgIDAndUserID :one
SELECT * FROM organization_memberships
WHERE organization_id = $1 AND user_id = $2 AND status NOT IN ('left', 'removed');

-- name: GetOrganizationMembershipsByUserID :many
SELECT * FROM organization_memberships
WHERE user_id = $1 AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1 AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetOrganizationMembershipsByOrgIDAndRole :many
SELECT * FROM organization_memberships
WHERE organization_id = $1 
    AND role = $2 
    AND status NOT IN ('left', 'removed')
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- NOTE: This method returns ALL statuses including 'left' and 'removed'
-- Use for admin operations where you need to see deleted memberships
-- name: GetOrganizationMembershipsByOrgIDAndStatus :many
SELECT * FROM organization_memberships
WHERE organization_id = $1 AND status = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- NOTE: This method returns membership regardless of status
-- Use for operations that need to see left/removed memberships
-- name: GetOrganizationMembershipByIDWithAllStatuses :one
SELECT * FROM organization_memberships
WHERE id = $1;

-- ============================================
-- COUNT METHODS
-- ============================================

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
WHERE organization_id = $1 AND status = $2;

-- ============================================
-- UPDATE METHODS
-- ============================================

-- name: UpdateOrganizationMembershipRole :one
UPDATE organization_memberships
SET 
    role = $2,
    updated_at = NOW()
WHERE id = $1 
    AND status NOT IN ('left', 'removed')
RETURNING *;

-- name: UpdateOrganizationMembershipStatus :one
UPDATE organization_memberships
SET 
    status = $2,
    updated_at = NOW()
WHERE id = $1 
    AND status NOT IN ('left', 'removed')
RETURNING *;

-- NOTE: Use this when you need to update status without the active check
-- For example, when undoing a soft delete or updating banned->active
-- name: UpdateOrganizationMembershipStatusByID :one
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


-- ============================================
-- SOFT DELETE METHODS
-- ============================================

-- name: SoftDeleteOrganizationMembership :exec
UPDATE organization_memberships
SET 
    status = 'left',
    deleted_by_mem_id = $2,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 
    AND status NOT IN ('left', 'removed');

-- name: SoftDeleteOrganizationMembershipByUserAndOrg :exec
UPDATE organization_memberships
SET 
    status = 'left',
    deleted_by_mem_id = $3,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1 
    AND organization_id = $2 
    AND status NOT IN ('left', 'removed');

-- NOTE: This sets status to 'removed' (terminal state, cannot be undone)
-- name: RemoveOrganizationMembership :exec
UPDATE organization_memberships
SET 
    status = 'removed',
    deleted_by_mem_id = $2,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 
    AND status NOT IN ('left', 'removed');

-- NOTE: Restore a soft-deleted membership (status 'left' -> 'active')
-- name: RestoreOrganizationMembership :one
UPDATE organization_memberships
SET 
    status = 'active',
    deleted_by_mem_id = NULL,
    deleted_at = NULL,
    updated_at = NOW()
WHERE id = $1 AND status = 'left'
RETURNING *;

-- ============================================
-- HARD DELETE METHODS (Use sparingly)
-- ============================================

-- name: HardDeleteOrganizationMembership :exec
-- WARNING: This permanently deletes the record. Use only for testing or cleanup.
DELETE FROM organization_memberships
WHERE id = $1;

-- name: HardDeleteOrganizationMembershipByOrgID :exec
-- WARNING: This permanently deletes all memberships for an organization.
DELETE FROM organization_memberships
WHERE organization_id = $1;

-- ============================================
-- BATCH METHODS
-- ============================================

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

-- ============================================
-- ADMIN METHODS (Include all statuses)
-- ============================================

-- NOTE: This returns ALL memberships including deleted ones (for admin audit)
-- name: GetAllOrganizationMembershipsByOrgID :many
SELECT * FROM organization_memberships
WHERE organization_id = $1
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- NOTE: Get history of a user's memberships (including left/removed)
-- name: GetOrganizationMembershipHistory :many
SELECT * FROM organization_memberships
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
