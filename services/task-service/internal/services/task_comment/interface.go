package taskcomment

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	taskcommentrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task_comment"
)

type TaskCommentService interface {
	CreateTaskComment(ctx context.Context, req *taskv1.CreateTaskCommentRequest, userInfo *dto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError)
	UpdateTaskComment(ctx context.Context, req *taskv1.UpdateTaskCommentRequest, userInfo *dto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError)
	DeleteTaskComment(ctx context.Context, req *taskv1.DeleteTaskCommentRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	ListTaskComments(ctx context.Context, req *taskv1.ListTaskCommentsRequest, userInfo *dto.UserInfo) (*taskv1.ListTaskCommentsResponse, *apperror.AppError)
}

type service struct {
	repo  taskcommentrepo.TaskCommentRepository
	authz authz.Authorizer
	tuples coreopenfga.TuppleManager
}

func NewTaskCommentService(repo taskcommentrepo.TaskCommentRepository, authz authz.Authorizer, tuples coreopenfga.TuppleManager) (TaskCommentService, *apperror.AppError) {
	if repo == nil || authz == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize task comment service").WithDetail("repo", "task comment repository must be configured")
	}

	return &service{
		repo:  repo,
		authz: authz,
		tuples: tuples,
	}, nil
}
