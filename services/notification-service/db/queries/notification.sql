-- name: GetNotification :one
SELECT *
FROM notifications
WHERE id = $1 LIMIT 1;

-- name: GetNotificationsByUserID :many
SELECT *
FROM notifications
WHERE recepient_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetNotificationsByUserIDAndStatus :many
SELECT *
FROM notifications
WHERE status = $2 AND recepient_user_id = $1
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateNotificationStatus :exec
UPDATE notifications
SET status = $2
WHERE id = $1;

-- name: CreateNotification :one
INSERT INTO notifications (
    recepient_email,
    recepient_user_id,
    message_data,
    status,
    template_type,
    retry_count
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;


-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;

-- name: DeleteNotificationsByUserID :exec
DELETE FROM notifications
WHERE recepient_user_id = $1;

-- NOTE: these are not for user case

-- name: GetNotificationsByStatus :many
SELECT *
FROM notifications
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetNotifications :many
SELECT *
FROM notifications
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteNotificationsByStatus :exec
DELETE FROM notifications
WHERE status = $1;

