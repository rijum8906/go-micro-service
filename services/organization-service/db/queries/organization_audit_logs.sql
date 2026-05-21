-- name: CreateOrganizationAuditLog :one
INSERT INTO organization_audit_logs (
    organization_id,
    actor_mem_id,
    action,
    target_type,
    target_id,
    old_value,
    new_value,
    metadata,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, COALESCE(sqlc.arg(metadata), '{}'::jsonb), $8, $9
)
RETURNING *;

-- name: GetOrganizationAuditLog :one
SELECT * FROM organization_audit_logs
WHERE id = $1
LIMIT 1;

-- name: GetOrganizationAuditLogsByOrgID :many
SELECT * FROM organization_audit_logs
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrganizationAuditLogsByActor :many
SELECT * FROM organization_audit_logs
WHERE organization_id = $1
  AND actor_mem_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrganizationAuditLogsByTarget :many
SELECT * FROM organization_audit_logs
WHERE organization_id = $1
  AND target_type = $2
  AND target_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetOrganizationAuditLogsByAction :many
SELECT * FROM organization_audit_logs
WHERE organization_id = $1
  AND action = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetRecentOrganizationAuditLogs :many
SELECT * FROM organization_audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteOrganizationAuditLogsByOrgID :exec
DELETE FROM organization_audit_logs
WHERE organization_id = $1;
