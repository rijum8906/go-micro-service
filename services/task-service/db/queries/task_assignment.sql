-- name: AssignTask :one
INSERT INTO task_assignments (
    task_id,
    assignee_type,
    assignee_id,
    assigned_by
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetActiveTaskAssignment :one
SELECT *
FROM task_assignments
WHERE task_id = $1
  AND assignee_type = $2
  AND assignee_id = $3
  AND unassigned_at IS NULL
LIMIT 1;

-- name: UnassignTask :one
UPDATE task_assignments
SET unassigned_at = now()
WHERE task_id = $1
  AND assignee_type = $2
  AND assignee_id = $3
  AND unassigned_at IS NULL
RETURNING *;

-- name: ListTaskAssignments :many
SELECT *
FROM task_assignments
WHERE task_id = $1
ORDER BY assigned_at DESC;

-- name: ListActiveAssignmentsByAssignee :many
SELECT *
FROM task_assignments
WHERE assignee_type = $1
  AND assignee_id = $2
  AND unassigned_at IS NULL
ORDER BY assigned_at DESC;
