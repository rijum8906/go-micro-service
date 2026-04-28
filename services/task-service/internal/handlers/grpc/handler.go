package grpc

import (
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	projectservice "github.com/rijum8906/relay/services/task-service/internal/services/project"
	taskservice "github.com/rijum8906/relay/services/task-service/internal/services/task"
)

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer
	projectService projectservice.ProjectService
	taskService    taskservice.TaskService
}

func NewTaskHandler(projectService projectservice.ProjectService, taskService taskservice.TaskService) *TaskHandler {
	return &TaskHandler{
		projectService: projectService,
		taskService:    taskService,
	}
}
