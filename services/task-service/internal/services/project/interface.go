package project

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	projectrepo "github.com/rijum8906/relay/services/task-service/internal/repository/project"
)

type ProjectService interface {
	CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError)
	GetProject(ctx context.Context, req *taskv1.GetProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError)
	UpdateProject(ctx context.Context, req *taskv1.UpdateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError)
	CompleteProject(ctx context.Context, req *taskv1.CompleteProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError)
	ArchiveProject(ctx context.Context, req *taskv1.ArchiveProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError)
	DeleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	ListProjects(ctx context.Context, req *taskv1.ListProjectsRequest, userInfo *dto.UserInfo) (*taskv1.ListProjectsResponse, *apperror.AppError)
}

type service struct {
	repo  projectrepo.ProjectRepository
	authz authz.Authorizer
}

func NewProjectService(repo projectrepo.ProjectRepository, authz authz.Authorizer) (ProjectService, *apperror.AppError) {
	if repo == nil || authz == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize project service").WithDetail("repo", "project repository must be configured")
	}

	return &service{
		repo:  repo,
		authz: authz}, nil
}
