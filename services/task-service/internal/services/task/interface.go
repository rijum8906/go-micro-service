package task

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type service struct {
	taskv1.UnimplementedTaskServiceServer

	q      db.Querier
	authz  authz.Authorizer
	tuples coreopenfga.TuppleManager
}

func New(q db.Querier, tuples coreopenfga.TuppleManager) (taskv1.TaskServiceServer, *apperror.AppError) {
	authorizer, appErr := authz.NewAuthorizer(q, tuples)
	if appErr != nil {
		return nil, appErr
	}

	return newService(q, authorizer, tuples)
}

func newService(q db.Querier, authorizer authz.Authorizer, tuples coreopenfga.TuppleManager) (*service, *apperror.AppError) {
	if q == nil || authorizer == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize task service").WithDetail("queries", "task queries must be configured")
	}

	return &service{
		q:      q,
		authz:  authorizer,
		tuples: tuples,
	}, nil
}
