package task

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (s *service) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("create task request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if strings.TrimSpace(req.GetTitle()) == "" {
		return nil, apperror.ErrValidation.WithMessage("title is required")
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	organizationID, appErr := utils.ParseOptionalUUID(req.GetOrganizationId(), "organization_id")
	if appErr != nil {
		return nil, appErr
	}

	projectID, appErr := utils.ParseOptionalUUID(req.GetProjectId(), "project_id")
	if appErr != nil {
		return nil, appErr
	}

	parentTaskID, appErr := utils.ParseOptionalUUID(req.GetParentTaskId(), "parent_task_id")
	if appErr != nil {
		return nil, appErr
	}

	dueAt, appErr := utils.ParseOptionalTimestamp(req.GetDueAt(), "due_at")
	if appErr != nil {
		return nil, appErr
	}

	priority := req.GetPriority()
	if priority == "" {
		priority = "medium"
	}

	task, appErr := s.repo.CreateTask(ctx, db.CreateTaskParams{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		ParentTaskID:   parentTaskID,
		CreatedBy:      userID,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       priority,
		DueAt:          dueAt,
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get task request is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("task id is required")
	}

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}

	task, appErr := s.repo.GetTask(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list tasks by project request is required")
	}
	if strings.TrimSpace(req.GetProjectId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}

	projectUUID, appErr := utils.NewUUID(req.GetProjectId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "project_id")
	}

	tasks, appErr := s.repo.ListTasksByProject(ctx, pgtype.UUID{
		Bytes: projectUUID,
		Valid: true,
	})
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTasksByProjectResponse{
		Tasks: mapTasks(tasks),
	}, nil
}
