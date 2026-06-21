package task

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (s *service) createTask(ctx context.Context, req *taskv1.CreateTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
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
	var parentTask *db.Task
	if parentTaskID.Valid {
		parentTask, appErr = s.authz.RequireTaskPermission(ctx, parentTaskID.Bytes, userInfo, taskpermissions.PermissionCanView)
		if appErr != nil {
			return nil, appErr
		}
		if appErr = validateChildTaskScope(parentTask, organizationID, projectID); appErr != nil {
			return nil, appErr
		}
	}
	if projectID.Valid {
		if _, appErr = s.authz.RequireProjectPermission(ctx, projectID.Bytes, userInfo, taskpermissions.PermissionCanContributeTasks); appErr != nil {
			return nil, appErr
		}
	}

	dueAt, appErr := utils.ParseOptionalTimestamp(req.GetDueAt(), "due_at")
	if appErr != nil {
		return nil, appErr
	}

	priority, appErr := normalizeTaskPriority(req.GetPriority(), "medium")
	if appErr != nil {
		return nil, appErr
	}

	taskRow, err := s.q.CreateTask(ctx, db.CreateTaskParams{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		ParentTaskID:   parentTaskID,
		CreatedBy:      userID,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       priority,
		DueAt:          dueAt,
	})
	task, appErr := utils.QueryOne(taskRow, err, "", "failed to create task")
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		writes := []client.ClientTupleKey{
			authz.TaskCreatorTuple(task.ID, task.CreatedBy),
		}
		if task.ProjectID.Valid {
			writes = append(writes, authz.TaskProjectTuple(task.ID, task.ProjectID.Bytes))
		}
		if appErr := s.tuples.Write(ctx, writes); appErr != nil {
			return nil, appErr
		}
	}

	return mapTask(task), nil
}

func (s *service) getTask(ctx context.Context, req *taskv1.GetTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get task request is required")
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("task id is required")
	}

	userIDCheck, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userIDCheck

	id, appErr := utils.NewUUID(req.GetId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "id")
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, id, userInfo, taskpermissions.PermissionCanView); appErr != nil {
		return nil, appErr
	}

	taskRow, err := s.q.GetTask(ctx, id)
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to get task")
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) listTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest, userInfo *dto.UserInfo) (*taskv1.ListTasksByProjectResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list tasks by project request is required")
	}
	if strings.TrimSpace(req.GetProjectId()) == "" {
		return nil, apperror.ErrValidation.WithMessage("project id is required")
	}

	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	status, appErr := validateOptionalTaskStatus(req.GetStatus())
	if appErr != nil {
		return nil, appErr
	}

	projectUUID, appErr := utils.NewUUID(req.GetProjectId())
	if appErr != nil {
		return nil, appErr.WithDetail("field", "project_id")
	}
	if _, appErr = s.authz.RequireProjectPermission(ctx, projectUUID, userInfo, taskpermissions.PermissionCanView); appErr != nil {
		return nil, appErr
	}

	taskRows, err := s.q.ListTasksByProject(ctx, pgtype.UUID{
		Bytes: projectUUID,
		Valid: true,
	})
	tasks, appErr := utils.QueryMany(taskRows, err, "failed to list project tasks")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTasksByProjectResponse{
		Tasks: mapTasks(filterTasksByStatus(tasks, status)),
	}, nil
}

func (s *service) updateTask(ctx context.Context, req *taskv1.UpdateTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update task request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if strings.TrimSpace(req.GetTitle()) == "" {
		return nil, apperror.ErrValidation.WithMessage("title is required")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, id, userInfo, taskpermissions.PermissionCanEdit); appErr != nil {
		return nil, appErr
	}

	existingTaskRow, err := s.q.GetTask(ctx, id)
	existingTask, appErr := utils.QueryOne(existingTaskRow, err, "task not found", "failed to get task")
	if appErr != nil {
		return nil, appErr
	}

	updatedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	dueAt, appErr := utils.ParseOptionalTimestamp(req.GetDueAt(), "due_at")
	if appErr != nil {
		return nil, appErr
	}

	priority, appErr := normalizeTaskPriority(req.GetPriority(), existingTask.Priority)
	if appErr != nil {
		return nil, appErr
	}

	taskRow, err := s.q.UpdateTask(ctx, db.UpdateTaskParams{
		ID:          id,
		UpdatedBy:   utils.PGUUID(updatedBy),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Priority:    priority,
		DueAt:       dueAt,
	})
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to update task")
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) deleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("delete task request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, id, userInfo, taskpermissions.PermissionCanDelete); appErr != nil {
		return nil, appErr
	}

	updatedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	deletedTaskRow, err := s.q.DeleteTask(ctx, db.DeleteTaskParams{
		ID:        id,
		UpdatedBy: utils.PGUUID(updatedBy),
	})
	if _, appErr = utils.QueryOne(deletedTaskRow, err, "task not found", "failed to delete task"); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) archiveTask(ctx context.Context, req *taskv1.ArchiveTaskRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("archive task request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, id, userInfo, taskpermissions.PermissionCanArchive); appErr != nil {
		return nil, appErr
	}

	updatedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	taskRow, err := s.q.ArchiveTask(ctx, db.ArchiveTaskParams{
		ID:        id,
		UpdatedBy: utils.PGUUID(updatedBy),
	})
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to archive task")
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) updateTaskStatus(ctx context.Context, req *taskv1.UpdateTaskStatusRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update task status request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, id, userInfo, taskpermissions.PermissionCanUpdateStatus); appErr != nil {
		return nil, appErr
	}

	status, appErr := validateTaskStatus(req.GetStatus())
	if appErr != nil {
		return nil, appErr
	}

	existingTaskRow, err := s.q.GetTask(ctx, id)
	existingTask, appErr := utils.QueryOne(existingTaskRow, err, "task not found", "failed to get task")
	if appErr != nil {
		return nil, appErr
	}

	updatedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	startedAt := existingTask.StartedAt
	completedAt := existingTask.CompletedAt
	now := time.Now()

	switch status {
	case "pending":
		startedAt = pgtype.Timestamptz{}
		completedAt = pgtype.Timestamptz{}
	case "in_progress":
		if !startedAt.Valid {
			startedAt = timestamptz(now)
		}
		completedAt = pgtype.Timestamptz{}
	case "completed":
		if !startedAt.Valid {
			startedAt = timestamptz(now)
		}
		completedAt = timestamptz(now)
	default:
		completedAt = pgtype.Timestamptz{}
	}

	taskRow, err := s.q.UpdateTaskStatus(ctx, db.UpdateTaskStatusParams{
		ID:          id,
		UpdatedBy:   utils.PGUUID(updatedBy),
		Status:      status,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
	})
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to update task status")
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) updateTaskProgress(ctx context.Context, req *taskv1.UpdateTaskProgressRequest, userInfo *dto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update task progress request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if req.GetProgressPercent() < 0 || req.GetProgressPercent() > 100 {
		return nil, apperror.ErrValidation.WithMessage("progress_percent must be between 0 and 100")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, id, userInfo, taskpermissions.PermissionCanUpdateProgress); appErr != nil {
		return nil, appErr
	}

	updatedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	taskRow, err := s.q.UpdateTaskProgress(ctx, db.UpdateTaskProgressParams{
		ID:              id,
		UpdatedBy:       utils.PGUUID(updatedBy),
		ProgressPercent: int16(req.GetProgressPercent()),
	})
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to update task progress")
	if appErr != nil {
		return nil, appErr
	}

	return mapTask(task), nil
}

func (s *service) listTasksByOrganization(ctx context.Context, req *taskv1.ListTasksByOrganizationRequest, userInfo *dto.UserInfo) (*taskv1.ListTasksByOrganizationResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list tasks by organization request is required")
	}
	if _, appErr := utils.ValidateUserInfo(userInfo); appErr != nil {
		return nil, appErr
	}

	organizationID, appErr := requiredUUID(req.GetOrganizationId(), "organization_id", "organization id is required")
	if appErr != nil {
		return nil, appErr
	}

	if _, appErr = validateOptionalTaskStatus(req.GetStatus()); appErr != nil {
		return nil, appErr
	}

	taskRows, err := s.q.ListTasksByOrganization(ctx, db.ListTasksByOrganizationParams{
		OrganizationID: utils.PGUUID(organizationID),
		Column2:        req.GetStatus(),
	})
	tasks, appErr := utils.QueryMany(taskRows, err, "failed to list organization tasks")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTasksByOrganizationResponse{
		Tasks: mapTasks(tasks),
	}, nil
}

func (s *service) listTasksByParent(ctx context.Context, req *taskv1.ListTasksByParentRequest, userInfo *dto.UserInfo) (*taskv1.ListTasksByParentResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list tasks by parent request is required")
	}
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	parentTaskID, appErr := requiredUUID(req.GetParentTaskId(), "parent_task_id", "parent task id is required")
	if appErr != nil {
		return nil, appErr
	}
	parentTask, appErr := s.authz.RequireTaskPermission(ctx, parentTaskID, userInfo, taskpermissions.PermissionCanView)
	if appErr != nil {
		return nil, appErr
	}

	taskRows, err := s.q.ListTasksByParent(ctx, utils.PGUUID(parentTaskID))
	tasks, appErr := utils.QueryMany(taskRows, err, "failed to list child tasks")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTasksByParentResponse{
		Tasks: mapTasks(filterTasksByParentScope(tasks, parentTask, userID)),
	}, nil
}

func (s *service) listTasksByCreator(ctx context.Context, req *taskv1.ListTasksByCreatorRequest, userInfo *dto.UserInfo) (*taskv1.ListTasksByCreatorResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list tasks by creator request is required")
	}
	if _, appErr := utils.ValidateUserInfo(userInfo); appErr != nil {
		return nil, appErr
	}

	createdBy, appErr := requiredUUID(req.GetCreatedBy(), "created_by", "created_by is required")
	if appErr != nil {
		return nil, appErr
	}

	if _, appErr = validateOptionalTaskStatus(req.GetStatus()); appErr != nil {
		return nil, appErr
	}

	taskRows, err := s.q.ListTasksByCreator(ctx, db.ListTasksByCreatorParams{
		CreatedBy: createdBy,
		Column2:   req.GetStatus(),
	})
	tasks, appErr := utils.QueryMany(taskRows, err, "failed to list creator tasks")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTasksByCreatorResponse{
		Tasks: mapTasks(tasks),
	}, nil
}

func normalizeTaskPriority(value, fallback string) (string, *apperror.AppError) {
	priority := strings.TrimSpace(value)
	if priority == "" {
		priority = fallback
	}

	switch priority {
	case "low", "medium", "high", "urgent":
		return priority, nil
	default:
		return "", apperror.ErrValidation.WithMessage("invalid task priority").WithDetail("field", "priority")
	}
}

func validateOptionalTaskStatus(value string) (string, *apperror.AppError) {
	status := strings.TrimSpace(value)
	if status == "" {
		return "", nil
	}

	return validateTaskStatus(status)
}

func validateTaskStatus(value string) (string, *apperror.AppError) {
	status := strings.TrimSpace(value)
	switch status {
	case "pending", "in_progress", "blocked", "completed", "cancelled":
		return status, nil
	default:
		return "", apperror.ErrValidation.WithMessage("invalid task status").WithDetail("field", "status")
	}
}

func filterTasksByStatus(tasks []db.Task, status string) []db.Task {
	if _, appErr := validateOptionalTaskStatus(status); appErr != nil {
		return tasks
	}
	if strings.TrimSpace(status) == "" {
		return tasks
	}

	filtered := make([]db.Task, 0, len(tasks))
	for i := range tasks {
		if tasks[i].Status == status {
			filtered = append(filtered, tasks[i])
		}
	}

	return filtered
}

func validateChildTaskScope(parent *db.Task, organizationID, projectID pgtype.UUID) *apperror.AppError {
	if parent == nil {
		return nil
	}

	if parent.ProjectID.Valid {
		if !optionalUUIDEqual(parent.ProjectID, projectID) || organizationID.Valid {
			return invalidParentTaskScope()
		}
		return nil
	}

	if projectID.Valid {
		return invalidParentTaskScope()
	}

	if !optionalUUIDEqual(parent.OrganizationID, organizationID) {
		return invalidParentTaskScope()
	}

	return nil
}

func filterTasksByParentScope(tasks []db.Task, parent *db.Task, userID uuid.UUID) []db.Task {
	filtered := make([]db.Task, 0, len(tasks))
	for i := range tasks {
		if taskMatchesParentScope(&tasks[i], parent, userID) {
			filtered = append(filtered, tasks[i])
		}
	}
	return filtered
}

func taskMatchesParentScope(task, parent *db.Task, userID uuid.UUID) bool {
	if task == nil || parent == nil {
		return false
	}

	if parent.ProjectID.Valid {
		return optionalUUIDEqual(task.ProjectID, parent.ProjectID) && !task.OrganizationID.Valid
	}

	if task.ProjectID.Valid || task.CreatedBy != userID {
		return false
	}

	return optionalUUIDEqual(task.OrganizationID, parent.OrganizationID)
}

func optionalUUIDEqual(left, right pgtype.UUID) bool {
	if left.Valid != right.Valid {
		return false
	}
	if !left.Valid {
		return true
	}
	return left.Bytes == right.Bytes
}

func invalidParentTaskScope() *apperror.AppError {
	return apperror.ErrValidation.WithMessage("child task scope must match parent task scope").WithDetail("field", "parent_task_id")
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}
