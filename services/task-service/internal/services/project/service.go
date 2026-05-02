package project

import (
	"context"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
	"strings"
)

func (s *service) CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
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

	project, appErr := s.repo.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: organizationID,
		CreatedBy:      createdBy,
		Name:           req.GetName(),
		Description:    req.GetDescription(),
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) GetProject(ctx context.Context, req *taskv1.GetProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
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

	project, appErr := s.repo.GetProject(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) UpdateProject(ctx context.Context, req *taskv1.UpdateProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
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

	project, appErr := s.repo.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          id,
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) CompleteProject(ctx context.Context, req *taskv1.CompleteProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
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

	project, appErr := s.repo.CompleteProject(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) ArchiveProject(ctx context.Context, req *taskv1.ArchiveProjectRequest, userInfo *dto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
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

	project, appErr := s.repo.ArchiveProject(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	return mapProject(project), nil
}

func (s *service) DeleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
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

	_, appErr = s.repo.DeleteProject(ctx, db.DeleteProjectParams{
		ID:        id,
		DeletedBy: utils.PGUUID(deletedBy),
	})
	if appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) ListProjects(ctx context.Context, req *taskv1.ListProjectsRequest, userInfo *dto.UserInfo) (*taskv1.ListProjectsResponse, *apperror.AppError) {
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

	projects, appErr := s.repo.ListProjects(ctx, db.ListProjectsParams{
		OrganizationID: organizationID,
		Status:         req.GetStatus(),
	})
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListProjectsResponse{
		Projects: mapProjects(projects),
	}, nil
}
