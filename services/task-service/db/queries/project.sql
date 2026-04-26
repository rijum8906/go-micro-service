-- name: CreateProject :one
INSERT INTO projects (
    organization_id,
    created_by,
    name,
    description
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetProject :one
SELECT *
FROM projects
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateProject :one
UPDATE projects
SET name = $2,
    description = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: CompleteProject :one
UPDATE projects
SET status = 'completed',
    completed_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveProject :one
UPDATE projects
SET status = 'archived',
    archived_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProject :one
UPDATE projects
SET deleted_at = now(),
    deleted_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ListProjects :many
SELECT *
FROM projects
WHERE deleted_at IS NULL
  AND (
    ($1::uuid IS NULL AND organization_id IS NULL)
    OR organization_id = $1
  )
  AND ($2::text = '' OR status = $2)
ORDER BY created_at DESC;
