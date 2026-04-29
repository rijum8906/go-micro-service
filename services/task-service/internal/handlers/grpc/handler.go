package grpc

import (
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	projectservice "github.com/rijum8906/relay/services/task-service/internal/services/project"
	projectmembershipservice "github.com/rijum8906/relay/services/task-service/internal/services/project_membership"
	taskservice "github.com/rijum8906/relay/services/task-service/internal/services/task"
	taskassigmentservice "github.com/rijum8906/relay/services/task-service/internal/services/task_assigment"
	taskcommentservice "github.com/rijum8906/relay/services/task-service/internal/services/task_comment"
)

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer
	projectService           projectservice.ProjectService
	projectMembershipService projectmembershipservice.ProjectMembershipService
	taskService              taskservice.TaskService
	taskAssignmentService    taskassigmentservice.TaskAssignmentService
	taskCommentService       taskcommentservice.TaskCommentService
}

func NewTaskHandler(
	projectService projectservice.ProjectService,
	projectMembershipService projectmembershipservice.ProjectMembershipService,
	taskService taskservice.TaskService,
	taskAssignmentService taskassigmentservice.TaskAssignmentService,
	taskCommentService taskcommentservice.TaskCommentService,
) *TaskHandler {
	return &TaskHandler{
		projectService:           projectService,
		projectMembershipService: projectMembershipService,
		taskService:              taskService,
		taskAssignmentService:    taskAssignmentService,
		taskCommentService:       taskCommentService,
	}
}
