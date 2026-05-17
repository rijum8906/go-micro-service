package task

import (
	"context"
	"strings"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	taskpermissions "github.com/rijum8906/relay/packages/core/permissions/task"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (s *service) createProject(ctx context.Context, req *taskv1.CreateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("create project request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, apperror.ErrValidation.WithMessage("name is required")
	}

	createdBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	organizationID, appErr := utils.ParseOptionalUUID(req.GetOrganizationId(), "organization_id")
	if appErr != nil {
		return nil, appErr
	}

	projectRow, err := s.q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: organizationID,
		CreatedBy:      createdBy,
		Name:           req.GetName(),
		Description:    req.GetDescription(),
	})
	project, appErr := utils.QueryOne(projectRow, err, "", "failed to create project")
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Write(ctx, []client.ClientTupleKey{
			authz.ProjectRoleTuple(project.ID, string(taskpermissions.RoleOwner), createdBy),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return mapProject(project), nil
}

func (s *service) getProject(ctx context.Context, req *taskv1.GetProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get project request is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}

	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}

	if _, appErr := s.authz.RequireProjectRole(ctx, id, userInfo, authz.RoleMember); appErr != nil {
		return nil, appErr
	}

	projectRow, err := s.q.GetProject(ctx, id)
	project, appErr := utils.QueryOne(projectRow, err, "project not found", "failed to get project")
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) updateProject(ctx context.Context, req *taskv1.UpdateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update project request is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, apperror.ErrValidation.WithMessage("name is required")
	}

	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}

	if _, appErr := s.authz.RequireProjectRole(ctx, id, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	projectRow, err := s.q.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          id,
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	project, appErr := utils.QueryOne(projectRow, err, "project not found", "failed to update project")
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) completeProject(ctx context.Context, req *taskv1.CompleteProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("complete project request is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}

	if _, appErr := s.authz.RequireProjectRole(ctx, id, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	projectRow, err := s.q.CompleteProject(ctx, id)
	project, appErr := utils.QueryOne(projectRow, err, "project not found", "failed to complete project")
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) archiveProject(ctx context.Context, req *taskv1.ArchiveProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("archive project request is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}

	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}

	if _, appErr := s.authz.RequireProjectRole(ctx, id, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	projectRow, err := s.q.ArchiveProject(ctx, id)
	project, appErr := utils.QueryOne(projectRow, err, "project not found", "failed to archive project")
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) deleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("delete project request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}

	deletedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	if _, appErr := s.authz.RequireProjectRole(ctx, id, userInfo, authz.RoleOwner); appErr != nil {
		return nil, appErr
	}

	deletedProject, err := s.q.DeleteProject(ctx, db.DeleteProjectParams{
		ID:        id,
		DeletedBy: utils.PGUUID(deletedBy),
	})
	_, appErr = utils.QueryOne(deletedProject, err, "project not found", "failed to delete project")
	if appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) listProjects(ctx context.Context, req *taskv1.ListProjectsRequest, userInfo *dto.UserInfo) (*taskv1.ListProjectsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list projects request is required")
	}
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	organizationID, appErr := utils.ParseOptionalUUID(req.GetOrganizationId(), "organization_id")
	if appErr != nil {
		return nil, appErr
	}

	projectRows, err := s.q.ListProjects(ctx, db.ListProjectsParams{
		OrganizationID: organizationID,
		Status:         req.GetStatus(),
	})
	projects, appErr := utils.QueryMany(projectRows, err, "failed to list projects")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListProjectsResponse{
		Projects: mapProjects(projects),
	}, nil
}
