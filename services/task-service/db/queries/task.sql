-- name: GetTask :one
SELECT *
FROM tasks
WHERE id = $1
LIMIT 1;

-- name: GetTasksByTeamID :many
SELECT *
FROM tasks
WHERE team_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTasksByAssignedID :many
SELECT *
FROM tasks
WHERE assigned_to_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTask :one
INSERT INTO tasks (
  title,
  description,
  priority,
  due_at,
  created_by_user_id,
  assigned_to_user_id,
  team_id,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2,
    description = $3,
    priority = $4,
    due_at = $5,
    assigned_to_user_id = $6,
    team_id = $7,
    status = $8,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;