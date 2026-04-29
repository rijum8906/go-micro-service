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
