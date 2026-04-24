package task

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

type TaskService interface {
	CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError)
	GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, *apperror.AppError)
	ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, *apperror.AppError)
}

type service struct {
	repos *utils.Repos
}

func NewTaskService(repos *utils.Repos) (TaskService, *apperror.AppError) {
	if repos == nil || repos.Task == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize task service").WithDetail("repos", "task repository is not configured")
	}

	return &service{repos: repos}, nil
}
