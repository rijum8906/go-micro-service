package projectmembership

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	projectmembershiprepo "github.com/rijum8906/relay/services/task-service/internal/repository/project_membership"
)

type ProjectMembershipService interface {
	AddProjectMember(ctx context.Context, req *taskv1.AddProjectMemberRequest, userInfo *dto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError)
	RemoveProjectMember(ctx context.Context, req *taskv1.RemoveProjectMemberRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	UpdateProjectMemberRole(ctx context.Context, req *taskv1.UpdateProjectMemberRoleRequest, userInfo *dto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError)
	ListProjectMembers(ctx context.Context, req *taskv1.ListProjectMembersRequest, userInfo *dto.UserInfo) (*taskv1.ListProjectMembersResponse, *apperror.AppError)
}

type service struct {
	repo  projectmembershiprepo.ProjectMembershipRepository
	authz authz.Authorizer
}

func NewProjectMembershipService(repo projectmembershiprepo.ProjectMembershipRepository, authz authz.Authorizer) (ProjectMembershipService, *apperror.AppError) {
	if repo == nil || authz == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize project membership service").WithDetail("repo", "project membership repository must be configured")
	}

	return &service{
		repo:  repo,
		authz: authz,
	}, nil
}
