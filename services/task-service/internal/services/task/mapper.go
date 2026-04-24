package task

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

func mapTask(task *db.Task) *modelsv1.Task {
	if task == nil {
		return nil
	}

	return &modelsv1.Task{
		Id:              task.ID.String(),
		OrganizationId:  uuidString(task.OrganizationID),
		ProjectId:       uuidString(task.ProjectID),
		ParentTaskId:    uuidString(task.ParentTaskID),
		CreatedBy:       task.CreatedBy.String(),
		UpdatedBy:       uuidString(task.UpdatedBy),
		Title:           task.Title,
		Description:     task.Description,
		Status:          task.Status,
		Priority:        task.Priority,
		ProgressPercent: int32(task.ProgressPercent),
		StartedAt:       timestamp(task.StartedAt),
		DueAt:           timestamp(task.DueAt),
		CompletedAt:     timestamp(task.CompletedAt),
		ArchivedAt:      timestamp(task.ArchivedAt),
		DeletedAt:       timestamp(task.DeletedAt),
		DeletedBy:       uuidString(task.DeletedBy),
		CreatedAt:       timestamp(task.CreatedAt),
		UpdatedAt:       timestamp(task.UpdatedAt),
	}
}

func mapTasks(tasks []db.Task) []*modelsv1.Task {
	items := make([]*modelsv1.Task, 0, len(tasks))
	for i := range tasks {
		items = append(items, mapTask(&tasks[i]))
	}
	return items
}

func uuidString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}
