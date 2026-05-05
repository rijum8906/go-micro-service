package taskassigment

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (s *service) AssignTask(ctx context.Context, req *taskv1.AssignTaskRequest, userInfo *dto.UserInfo) (*modelsv1.TaskAssignment, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("assign task request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	taskID, assigneeType, assigneeID, appErr := parseAssignment(req.GetTaskId(), req.GetAssigneeType(), req.GetAssigneeId(), "task_id", "assignee_id")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskRole(ctx, taskID, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	if _, appErr = s.repo.GetActiveTaskAssignment(ctx, db.GetActiveTaskAssignmentParams{
		TaskID:       taskID,
		AssigneeType: assigneeType,
		AssigneeID:   assigneeID,
	}); appErr == nil {
		return nil, apperror.ErrConflict.WithMessage("task assignment already exists")
	} else if appErr.Code != apperror.CodeNotFound {
		return nil, appErr
	}

	assignedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	assignment, appErr := s.repo.AssignTask(ctx, db.AssignTaskParams{
		TaskID:       taskID,
		AssigneeType: assigneeType,
		AssigneeID:   assigneeID,
		AssignedBy:   assignedBy,
	})
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Write(ctx, []client.ClientTupleKey{
			authz.TaskAssigneeTuple(taskID, assigneeType, assigneeID),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return mapTaskAssignment(assignment), nil
}

func (s *service) UnassignTask(ctx context.Context, req *taskv1.UnassignTaskRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("unassign task request is required")
	}
	if _, appErr := utils.ValidateUserInfo(userInfo); appErr != nil {
		return nil, appErr
	}

	taskID, assigneeType, assigneeID, appErr := parseAssignment(req.GetTaskId(), req.GetAssigneeType(), req.GetAssigneeId(), "task_id", "assignee_id")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskRole(ctx, taskID, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	assignment, appErr := s.repo.UnassignTask(ctx, db.UnassignTaskParams{
		TaskID:       taskID,
		AssigneeType: assigneeType,
		AssigneeID:   assigneeID,
	})
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			authz.DeleteTuple(authz.TaskAssigneeTuple(assignment.TaskID, assignment.AssigneeType, assignment.AssigneeID)),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) ReassignTask(ctx context.Context, req *taskv1.ReassignTaskRequest, userInfo *dto.UserInfo) (*modelsv1.TaskAssignment, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("reassign task request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	taskID, fromAssigneeType, fromAssigneeID, appErr := parseAssignment(req.GetTaskId(), req.GetFromAssigneeType(), req.GetFromAssigneeId(), "task_id", "from_assignee_id")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskRole(ctx, taskID, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	_, toAssigneeType, toAssigneeID, appErr := parseAssignment(req.GetTaskId(), req.GetToAssigneeType(), req.GetToAssigneeId(), "task_id", "to_assignee_id")
	if appErr != nil {
		return nil, appErr
	}

	if fromAssigneeType == toAssigneeType && fromAssigneeID == toAssigneeID {
		return nil, apperror.ErrValidation.WithMessage("reassignment target must be different from the current assignee")
	}

	fromParams := db.GetActiveTaskAssignmentParams{
		TaskID:       taskID,
		AssigneeType: fromAssigneeType,
		AssigneeID:   fromAssigneeID,
	}
	if _, appErr = s.repo.GetActiveTaskAssignment(ctx, fromParams); appErr != nil {
		return nil, appErr
	}

	toParams := db.GetActiveTaskAssignmentParams{
		TaskID:       taskID,
		AssigneeType: toAssigneeType,
		AssigneeID:   toAssigneeID,
	}
	if _, appErr = s.repo.GetActiveTaskAssignment(ctx, toParams); appErr == nil {
		return nil, apperror.ErrConflict.WithMessage("task assignment already exists")
	} else if appErr.Code != apperror.CodeNotFound {
		return nil, appErr
	}

	assignedBy, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	assignment, appErr := s.repo.AssignTask(ctx, db.AssignTaskParams{
		TaskID:       taskID,
		AssigneeType: toAssigneeType,
		AssigneeID:   toAssigneeID,
		AssignedBy:   assignedBy,
	})
	if appErr != nil {
		return nil, appErr
	}

	if _, appErr = s.repo.UnassignTask(ctx, db.UnassignTaskParams{
		TaskID:       taskID,
		AssigneeType: fromAssigneeType,
		AssigneeID:   fromAssigneeID,
	}); appErr != nil {
		_, _ = s.repo.UnassignTask(ctx, db.UnassignTaskParams{
			TaskID:       taskID,
			AssigneeType: toAssigneeType,
			AssigneeID:   toAssigneeID,
		})
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			authz.DeleteTuple(authz.TaskAssigneeTuple(taskID, fromAssigneeType, fromAssigneeID)),
		}); appErr != nil {
			return nil, appErr
		}
		if appErr := s.tuples.Write(ctx, []client.ClientTupleKey{
			authz.TaskAssigneeTuple(taskID, toAssigneeType, toAssigneeID),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return mapTaskAssignment(assignment), nil
}

func (s *service) ListTaskAssignments(ctx context.Context, req *taskv1.ListTaskAssignmentsRequest, userInfo *dto.UserInfo) (*taskv1.ListTaskAssignmentsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list task assignments request is required")
	}
	if _, appErr := utils.ValidateUserInfo(userInfo); appErr != nil {
		return nil, appErr
	}

	taskID, appErr := requiredUUID(req.GetTaskId(), "task_id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskRole(ctx, taskID, userInfo, authz.RoleMember); appErr != nil {
		return nil, appErr
	}

	assignments, appErr := s.repo.ListTaskAssignments(ctx, taskID)
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTaskAssignmentsResponse{
		Assignments: mapTaskAssignments(assignments),
	}, nil
}

func parseAssignment(taskIDValue, assigneeTypeValue, assigneeIDValue, taskField, assigneeField string) (uuid.UUID, string, uuid.UUID, *apperror.AppError) {
	taskID, appErr := requiredUUID(taskIDValue, taskField, "task id is required")
	if appErr != nil {
		return uuid.UUID{}, "", uuid.UUID{}, appErr
	}

	assigneeType, appErr := validateAssigneeType(assigneeTypeValue)
	if appErr != nil {
		return uuid.UUID{}, "", uuid.UUID{}, appErr
	}

	assigneeID, appErr := requiredUUID(assigneeIDValue, assigneeField, "assignee id is required")
	if appErr != nil {
		return uuid.UUID{}, "", uuid.UUID{}, appErr
	}

	return taskID, assigneeType, assigneeID, nil
}

func requiredUUID(value, field, requiredMessage string) (uuid.UUID, *apperror.AppError) {
	if strings.TrimSpace(value) == "" {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage(requiredMessage)
	}

	id, appErr := utils.NewUUID(value)
	if appErr != nil {
		return uuid.UUID{}, appErr.WithDetail("field", field)
	}

	return id, nil
}

func validateAssigneeType(value string) (string, *apperror.AppError) {
	assigneeType := strings.TrimSpace(value)
	switch assigneeType {
	case "user", "team":
		return assigneeType, nil
	default:
		return "", apperror.ErrValidation.WithMessage("invalid assignee type").WithDetail("field", "assignee_type")
	}
}
