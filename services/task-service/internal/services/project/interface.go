package project

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	projectrepo "github.com/rijum8906/relay/services/task-service/internal/repository/project"
)

type ProjectService interface {
	CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError)
	GetProject(ctx context.Context, req *taskv1.GetProjectRequest) (*modelsv1.Project, *apperror.AppError)
	UpdateProject(ctx context.Context, req *taskv1.UpdateProjectRequest) (*modelsv1.Project, *apperror.AppError)
	CompleteProject(ctx context.Context, req *taskv1.CompleteProjectRequest) (*modelsv1.Project, *apperror.AppError)
	ArchiveProject(ctx context.Context, req *taskv1.ArchiveProjectRequest) (*modelsv1.Project, *apperror.AppError)
	DeleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	ListProjects(ctx context.Context, req *taskv1.ListProjectsRequest) (*taskv1.ListProjectsResponse, *apperror.AppError)
}

type service struct {
	repo projectrepo.ProjectRepository
}

func NewProjectService(repo projectrepo.ProjectRepository) (ProjectService, *apperror.AppError) {
	if repo == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize project service").WithDetail("repo", "project repository must be configured")
	}

	return &service{repo: repo}, nil
}
