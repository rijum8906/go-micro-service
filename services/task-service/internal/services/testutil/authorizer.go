package testutil

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type allowAuthorizer struct{}

func NewAllowAuthorizer() authz.Authorizer {
	return allowAuthorizer{}
}

func (allowAuthorizer) RequireProjectRole(_ context.Context, projectID uuid.UUID, _ *coredto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
	return &db.ProjectMembership{
		ID:        uuid.New(),
		ProjectID: projectID,
		Role:      string(minRole),
	}, nil
}

func (allowAuthorizer) RequireTaskRole(_ context.Context, taskID uuid.UUID, userInfo *coredto.UserInfo, _ authz.Role) (*db.Task, *apperror.AppError) {
	task := &db.Task{ID: taskID}
	if userInfo != nil {
		if userID, err := uuid.Parse(userInfo.UserID); err == nil {
			task.CreatedBy = userID
		}
	}
	return task, nil
}
