-- name: CreateTaskComment :one
INSERT INTO task_comments (
    task_id,
    author_id,
    body
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetTaskComment :one
SELECT *
FROM task_comments
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateTaskComment :one
UPDATE task_comments
SET body = $2,
    edited_at = now(),
    edit_count = edit_count + 1
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTaskComment :one
UPDATE task_comments
SET deleted_at = now(),
    deleted_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ListTaskComments :many
SELECT *
FROM task_comments
WHERE task_id = $1
  AND deleted_at IS NULL
ORDER BY created_at ASC;
