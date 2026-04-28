package task

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func mapTask(task *db.Task) *modelsv1.Task {
	if task == nil {
		return nil
	}

	return &modelsv1.Task{
		Id:              task.ID.String(),
		OrganizationId:  utils.UUIDString(task.OrganizationID),
		ProjectId:       utils.UUIDString(task.ProjectID),
		ParentTaskId:    utils.UUIDString(task.ParentTaskID),
		CreatedBy:       task.CreatedBy.String(),
		UpdatedBy:       utils.UUIDString(task.UpdatedBy),
		Title:           task.Title,
		Description:     task.Description,
		Status:          task.Status,
		Priority:        task.Priority,
		ProgressPercent: int32(task.ProgressPercent),
		StartedAt:       utils.Timestamp(task.StartedAt),
		DueAt:           utils.Timestamp(task.DueAt),
		CompletedAt:     utils.Timestamp(task.CompletedAt),
		ArchivedAt:      utils.Timestamp(task.ArchivedAt),
		DeletedAt:       utils.Timestamp(task.DeletedAt),
		DeletedBy:       utils.UUIDString(task.DeletedBy),
		CreatedAt:       utils.Timestamp(task.CreatedAt),
		UpdatedAt:       utils.Timestamp(task.UpdatedAt),
	}
}

func mapTasks(tasks []db.Task) []*modelsv1.Task {
	items := make([]*modelsv1.Task, 0, len(tasks))
	for i := range tasks {
		items = append(items, mapTask(&tasks[i]))
	}
	return items
}
