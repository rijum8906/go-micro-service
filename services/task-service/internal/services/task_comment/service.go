package taskcomment

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (s *service) CreateTaskComment(ctx context.Context, req *taskv1.CreateTaskCommentRequest, userInfo *dto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("create task comment request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if strings.TrimSpace(req.GetBody()) == "" {
		return nil, apperror.ErrValidation.WithMessage("body is required")
	}

	taskID, appErr := requiredUUID(req.GetTaskId(), "task_id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}

	authorID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	comment, appErr := s.repo.CreateTaskComment(ctx, db.CreateTaskCommentParams{
		TaskID:   taskID,
		AuthorID: authorID,
		Body:     req.GetBody(),
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapTaskComment(comment), nil
}

func (s *service) UpdateTaskComment(ctx context.Context, req *taskv1.UpdateTaskCommentRequest, userInfo *dto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update task comment request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if strings.TrimSpace(req.GetBody()) == "" {
		return nil, apperror.ErrValidation.WithMessage("body is required")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "comment id is required")
	if appErr != nil {
		return nil, appErr
	}

	comment, appErr := s.repo.GetTaskComment(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}
	if comment.AuthorID != userID {
		return nil, apperror.ErrForbidden.WithMessage("only the comment author can update this comment")
	}

	comment, appErr = s.repo.UpdateTaskComment(ctx, db.UpdateTaskCommentParams{
		ID:   id,
		Body: req.GetBody(),
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapTaskComment(comment), nil
}

func (s *service) DeleteTaskComment(ctx context.Context, req *taskv1.DeleteTaskCommentRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("delete task comment request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	id, appErr := requiredUUID(req.GetId(), "id", "comment id is required")
	if appErr != nil {
		return nil, appErr
	}

	comment, appErr := s.repo.GetTaskComment(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}
	if comment.AuthorID != userID {
		return nil, apperror.ErrForbidden.WithMessage("only the comment author can delete this comment")
	}

	if _, appErr = s.repo.DeleteTaskComment(ctx, db.DeleteTaskCommentParams{
		ID:        id,
		DeletedBy: utils.PGUUID(userID),
	}); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) ListTaskComments(ctx context.Context, req *taskv1.ListTaskCommentsRequest) (*taskv1.ListTaskCommentsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list task comments request is required")
	}

	taskID, appErr := requiredUUID(req.GetTaskId(), "task_id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}

	comments, appErr := s.repo.ListTaskComments(ctx, taskID)
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTaskCommentsResponse{
		Comments: mapTaskComments(comments),
	}, nil
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
