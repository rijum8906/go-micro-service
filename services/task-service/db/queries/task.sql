-- name: CreateTask :one
INSERT INTO tasks (
    organization_id,
    project_id,
    parent_task_id,
    created_by,
    title,
    description,
    priority,
    due_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetTask :one
SELECT *
FROM tasks
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: ListTasksByProject :many
SELECT *
FROM tasks
WHERE project_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateTask :one
UPDATE tasks
SET updated_by = $2,
    title = $3,
    description = $4,
    priority = $5,
    due_at = $6
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTask :one
UPDATE tasks
SET updated_by = $2,
    deleted_at = now(),
    deleted_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveTask :one
UPDATE tasks
SET updated_by = $2,
    archived_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET updated_by = $2,
    status = $3,
    started_at = $4,
    completed_at = $5
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTaskProgress :one
UPDATE tasks
SET updated_by = $2,
    progress_percent = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ListTasksByOrganization :many
SELECT *
FROM tasks
WHERE organization_id = $1
  AND project_id IS NULL
  AND deleted_at IS NULL
  AND ($2::text = '' OR status = $2)
ORDER BY created_at DESC;

-- name: ListTasksByParent :many
SELECT *
FROM tasks
WHERE parent_task_id = $1
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: ListTasksByCreator :many
SELECT *
FROM tasks
WHERE created_by = $1
  AND deleted_at IS NULL
  AND ($2::text = '' OR status = $2)
ORDER BY created_at DESC;
