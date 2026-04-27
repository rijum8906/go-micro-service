package utils

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	projectRepo "github.com/rijum8906/relay/services/task-service/internal/repository/project"
	projectMembershipRepo "github.com/rijum8906/relay/services/task-service/internal/repository/project_membership"
	taskrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task"
	taskAssignment "github.com/rijum8906/relay/services/task-service/internal/repository/task_assignment"
	taskComment "github.com/rijum8906/relay/services/task-service/internal/repository/task_comment"
)

type Repos struct {
	Project           projectRepo.ProjectRepository
	Task              taskrepo.TaskRepository
	ProjectMembership projectMembershipRepo.ProjectMembershipRepository
	TaskAssignment    taskAssignment.TaskAssignmentRepository
	TaskComment       taskComment.TaskCommentRepository
}

func NewUUID(id string) (uuid.UUID, *apperror.AppError) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid uuid").WithDetail("error", err.Error())
	}
	return u, nil
}
