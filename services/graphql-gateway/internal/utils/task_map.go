package utils

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
)

func MapTask(task *modelsv1.Task) *model.Task {
	if task == nil {
		return nil
	}

	return &model.Task{
		ID:              task.Id,
		OrganizationID:  optionalString(task.OrganizationId),
		ProjectID:       optionalString(task.ProjectId),
		ParentTaskID:    optionalString(task.ParentTaskId),
		CreatedBy:       task.CreatedBy,
		UpdatedBy:       optionalString(task.UpdatedBy),
		Title:           task.Title,
		Description:     task.Description,
		Status:          task.Status,
		Priority:        task.Priority,
		ProgressPercent: task.ProgressPercent,
		StartedAt:       optionalTimestamp(task.StartedAt),
		DueAt:           optionalTimestamp(task.DueAt),
		CompletedAt:     optionalTimestamp(task.CompletedAt),
		ArchivedAt:      optionalTimestamp(task.ArchivedAt),
		DeletedAt:       optionalTimestamp(task.DeletedAt),
		DeletedBy:       optionalString(task.DeletedBy),
		CreatedAt:       requiredTimestamp(task.CreatedAt),
		UpdatedAt:       requiredTimestamp(task.UpdatedAt),
	}
}

func MapTasks(tasks []*modelsv1.Task) []*model.Task {
	mapped := make([]*model.Task, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		mapped = append(mapped, MapTask(task))
	}
	return mapped
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
