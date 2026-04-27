package projectmembership

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)


type  ProjectMembershipRepository interface {
	AddProjectMember(ctx context.Context, params db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError)
	GetActiveProjectMembership(ctx context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) 
	UpdateProjectMemberRole(ctx context.Context, params db.UpdateProjectMemberRoleParams)(*db.ProjectMembership, *apperror.AppError)
	RemoveProjectMember(ctx context.Context, params db.RemoveProjectMemberParams)(*db.ProjectMembership, *apperror.AppError)
	ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]db.ProjectMembership, *apperror.AppError)
	ListProjectMembershipsByUser(ctx context.Context, userID uuid.UUID)([]db.ProjectMembership, *apperror.AppError)
}