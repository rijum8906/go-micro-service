package projectmembership

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type projectMembershipRepository struct {
	q db.Querier
}

func NewProjectMembershipRepository(q db.Querier) ProjectMembershipRepository {
	return &projectMembershipRepository{q: q}
}

func(r *projectMembershipRepository) AddProjectMember(ctx context.Context, params db.AddProjectMemberParams)(*db.ProjectMembership, *apperror.AppError) {
	project, err := r.q.AddProjectMember(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to add project member").WithDetail("error", err.Error())		
	}
	
	return &project, nil
}

func(r *projectMembershipRepository) GetActiveProjectMembership(ctx context.Context, params db.GetActiveProjectMembershipParams)(*db.ProjectMembership, *apperror.AppError) {
	project, err := r.q.GetActiveProjectMembership(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to get active membership").WithDetail("error", err.Error())		
	}
	
	return &project, nil
}

func(r *projectMembershipRepository) UpdateProjectMemberRole(ctx context.Context, params db.UpdateProjectMemberRoleParams)(*db.ProjectMembership, *apperror.AppError) {
	project, err := r.q.UpdateProjectMemberRole(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to update project member role").WithDetail("error", err.Error())		
	}
	
	return &project, nil
}

func(r *projectMembershipRepository) RemoveProjectMember(ctx context.Context, params db.RemoveProjectMemberParams)(*db.ProjectMembership, *apperror.AppError) {
	project, err := r.q.RemoveProjectMember(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to remove project member").WithDetail("error", err.Error())		
	}
	
	return &project, nil
}

func(r *projectMembershipRepository) ListProjectMembers(ctx context.Context, projectID uuid.UUID)([]db.ProjectMembership, *apperror.AppError) {
	project, err := r.q.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list project members").WithDetail("error", err.Error())		
	}
	
	return project, nil
}

func(r *projectMembershipRepository) ListProjectMembershipsByUser(ctx context.Context, userID uuid.UUID)([]db.ProjectMembership, *apperror.AppError) {
	project, err := r.q.ListProjectMembershipsByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list project memberships by user").WithDetail("error", err.Error())		
	}
	
	return project, nil
}