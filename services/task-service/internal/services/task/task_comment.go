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

func (s *service) createTaskComment(ctx context.Context, req *taskv1.CreateTaskCommentRequest, userInfo *dto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError) {
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
	if _, appErr = s.authz.RequireTaskPermission(ctx, taskID, userInfo, taskpermissions.PermissionCanComment); appErr != nil {
		return nil, appErr
	}

	authorID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	commentRow, err := s.q.CreateTaskComment(ctx, db.CreateTaskCommentParams{
		TaskID:   taskID,
		AuthorID: authorID,
		Body:     req.GetBody(),
	})
	comment, appErr := utils.QueryOne(commentRow, err, "", "failed to create task comment")
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Write(ctx, []client.ClientTupleKey{
			authz.CommentAuthorTuple(comment.ID, authorID),
			authz.CommentTaskTuple(comment.ID, taskID),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return mapTaskComment(comment), nil
}

func (s *service) updateTaskComment(ctx context.Context, req *taskv1.UpdateTaskCommentRequest, userInfo *dto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError) {
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

	commentRow, err := s.q.GetTaskComment(ctx, id)
	comment, appErr := utils.QueryOne(commentRow, err, "task comment not found", "failed to get task comment")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, comment.TaskID, userInfo, taskpermissions.PermissionCanComment); appErr != nil {
		return nil, appErr
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}
	if comment.AuthorID != userID {
		return nil, apperror.ErrForbidden.WithMessage("only the comment author can update this comment")
	}

	commentRow, err = s.q.UpdateTaskComment(ctx, db.UpdateTaskCommentParams{
		ID:   id,
		Body: req.GetBody(),
	})
	comment, appErr = utils.QueryOne(commentRow, err, "task comment not found", "failed to update task comment")
	if appErr != nil {
		return nil, appErr
	}

	return mapTaskComment(comment), nil
}

func (s *service) deleteTaskComment(ctx context.Context, req *taskv1.DeleteTaskCommentRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
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

	commentRow, err := s.q.GetTaskComment(ctx, id)
	comment, appErr := utils.QueryOne(commentRow, err, "task comment not found", "failed to get task comment")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, comment.TaskID, userInfo, taskpermissions.PermissionCanComment); appErr != nil {
		return nil, appErr
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}
	if comment.AuthorID != userID {
		return nil, apperror.ErrForbidden.WithMessage("only the comment author can delete this comment")
	}

	deletedCommentRow, err := s.q.DeleteTaskComment(ctx, db.DeleteTaskCommentParams{
		ID:        id,
		DeletedBy: utils.PGUUID(userID),
	})
	if _, appErr = utils.QueryOne(deletedCommentRow, err, "task comment not found", "failed to delete task comment"); appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			authz.DeleteTuple(authz.CommentAuthorTuple(comment.ID, comment.AuthorID)),
			authz.DeleteTuple(authz.CommentTaskTuple(comment.ID, comment.TaskID)),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) listTaskComments(ctx context.Context, req *taskv1.ListTaskCommentsRequest, userInfo *dto.UserInfo) (*taskv1.ListTaskCommentsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list task comments request is required")
	}
	if _, appErr := utils.ValidateUserInfo(userInfo); appErr != nil {
		return nil, appErr
	}

	taskID, appErr := requiredUUID(req.GetTaskId(), "task_id", "task id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireTaskPermission(ctx, taskID, userInfo, taskpermissions.PermissionCanComment); appErr != nil {
		return nil, appErr
	}

	commentRows, err := s.q.ListTaskComments(ctx, taskID)
	comments, appErr := utils.QueryMany(commentRows, err, "failed to list task comments")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListTaskCommentsResponse{
		Comments: mapTaskComments(comments),
	}, nil
}
