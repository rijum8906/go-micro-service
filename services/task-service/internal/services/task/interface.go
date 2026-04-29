package task

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	taskrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task"
)

type TaskService interface {
	CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError)
	GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, *apperror.AppError)
	ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, *apperror.AppError)
	UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError)
	DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	ArchiveTask(ctx context.Context, req *taskv1.ArchiveTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError)
	UpdateTaskStatus(ctx context.Context, req *taskv1.UpdateTaskStatusRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError)
	UpdateTaskProgress(ctx context.Context, req *taskv1.UpdateTaskProgressRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError)
	ListTasksByOrganization(ctx context.Context, req *taskv1.ListTasksByOrganizationRequest) (*taskv1.ListTasksByOrganizationResponse, *apperror.AppError)
	ListTasksByParent(ctx context.Context, req *taskv1.ListTasksByParentRequest) (*taskv1.ListTasksByParentResponse, *apperror.AppError)
	ListTasksByCreator(ctx context.Context, req *taskv1.ListTasksByCreatorRequest) (*taskv1.ListTasksByCreatorResponse, *apperror.AppError)
}

type service struct {
	repo taskrepo.TaskRepository
}

func NewTaskService(repo taskrepo.TaskRepository) (TaskService, *apperror.AppError) {
	if repo == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize task service").WithDetail("repo", "task repository must be configured")
	}

	return &service{repo: repo}, nil
}
