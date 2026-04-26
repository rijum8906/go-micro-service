-- name: AddProjectMember :one
INSERT INTO project_memberships (
    project_id,
    user_id,
    role
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetActiveProjectMembership :one
SELECT *
FROM project_memberships
WHERE project_id = $1
  AND user_id = $2
  AND left_at IS NULL
LIMIT 1;

-- name: UpdateProjectMemberRole :one
UPDATE project_memberships
SET role = $3
WHERE project_id = $1
  AND user_id = $2
  AND left_at IS NULL
RETURNING *;

-- name: RemoveProjectMember :one
UPDATE project_memberships
SET left_at = now()
WHERE project_id = $1
  AND user_id = $2
  AND left_at IS NULL
RETURNING *;

-- name: ListProjectMembers :many
SELECT *
FROM project_memberships
WHERE project_id = $1
  AND left_at IS NULL
ORDER BY joined_at ASC;

-- name: ListProjectMembershipsByUser :many
SELECT *
FROM project_memberships
WHERE user_id = $1
  AND left_at IS NULL
ORDER BY joined_at DESC;
