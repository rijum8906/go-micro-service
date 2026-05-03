package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type stubProjectMembershipRepository struct{}

func (stubProjectMembershipRepository) AddProjectMember(context.Context, db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
	panic("unexpected AddProjectMember call")
}

func (stubProjectMembershipRepository) GetActiveProjectMembership(context.Context, db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
	panic("unexpected GetActiveProjectMembership call")
}

func (stubProjectMembershipRepository) UpdateProjectMemberRole(context.Context, db.UpdateProjectMemberRoleParams) (*db.ProjectMembership, *apperror.AppError) {
	panic("unexpected UpdateProjectMemberRole call")
}

func (stubProjectMembershipRepository) RemoveProjectMember(context.Context, db.RemoveProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
	panic("unexpected RemoveProjectMember call")
}

func (stubProjectMembershipRepository) ListProjectMembers(context.Context, uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
	panic("unexpected ListProjectMembers call")
}

func (stubProjectMembershipRepository) ListProjectMembershipsByUser(context.Context, uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
	panic("unexpected ListProjectMembershipsByUser call")
}

type stubTaskRepository struct {
	getTaskFn func(context.Context, uuid.UUID) (*db.Task, *apperror.AppError)
}

func (s stubTaskRepository) CreateTask(context.Context, db.CreateTaskParams) (*db.Task, *apperror.AppError) {
	panic("unexpected CreateTask call")
}

func (s stubTaskRepository) GetTask(ctx context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
	if s.getTaskFn == nil {
		panic("unexpected GetTask call")
	}
	return s.getTaskFn(ctx, id)
}

func (stubTaskRepository) ListTasksByProject(context.Context, pgtype.UUID) ([]db.Task, *apperror.AppError) {
	panic("unexpected ListTasksByProject call")
}

func (stubTaskRepository) UpdateTask(context.Context, db.UpdateTaskParams) (*db.Task, *apperror.AppError) {
	panic("unexpected UpdateTask call")
}

func (stubTaskRepository) DeleteTask(context.Context, db.DeleteTaskParams) (*db.Task, *apperror.AppError) {
	panic("unexpected DeleteTask call")
}

func (stubTaskRepository) ArchiveTask(context.Context, db.ArchiveTaskParams) (*db.Task, *apperror.AppError) {
	panic("unexpected ArchiveTask call")
}

func (stubTaskRepository) UpdateTaskStatus(context.Context, db.UpdateTaskStatusParams) (*db.Task, *apperror.AppError) {
	panic("unexpected UpdateTaskStatus call")
}

func (stubTaskRepository) UpdateTaskProgress(context.Context, db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError) {
	panic("unexpected UpdateTaskProgress call")
}

func (stubTaskRepository) ListTasksByOrganization(context.Context, db.ListTasksByOrganizationParams) ([]db.Task, *apperror.AppError) {
	panic("unexpected ListTasksByOrganization call")
}

func (stubTaskRepository) ListTasksByParent(context.Context, pgtype.UUID) ([]db.Task, *apperror.AppError) {
	panic("unexpected ListTasksByParent call")
}

func (stubTaskRepository) ListTasksByCreator(context.Context, db.ListTasksByCreatorParams) ([]db.Task, *apperror.AppError) {
	panic("unexpected ListTasksByCreator call")
}

func TestRequireTaskRoleAllowsCreatorForNonProjectTask(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()

	authorizer, appErr := NewAuthorizer(
		stubProjectMembershipRepository{},
		stubTaskRepository{
			getTaskFn: func(_ context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
				if id != taskID {
					t.Fatalf("unexpected task id: %s", id)
				}
				return &db.Task{ID: taskID, CreatedBy: userID}, nil
			},
		},
	)
	if appErr != nil {
		t.Fatalf("expected authorizer construction to succeed: %v", appErr)
	}

	task, appErr := authorizer.RequireTaskRole(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, RoleAdmin)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if task == nil || task.ID != taskID {
		t.Fatalf("unexpected task: %#v", task)
	}
}

func TestRequireTaskRoleRejectsNonCreatorForNonProjectTask(t *testing.T) {
	taskID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	authorizer, appErr := NewAuthorizer(
		stubProjectMembershipRepository{},
		stubTaskRepository{
			getTaskFn: func(_ context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
				if id != taskID {
					t.Fatalf("unexpected task id: %s", id)
				}
				return &db.Task{ID: taskID, CreatedBy: ownerID}, nil
			},
		},
	)
	if appErr != nil {
		t.Fatalf("expected authorizer construction to succeed: %v", appErr)
	}

	task, appErr := authorizer.RequireTaskRole(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, RoleMember)
	if task != nil {
		t.Fatalf("expected nil task, got %#v", task)
	}
	if appErr == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	if appErr.Code != apperror.CodeForbidden {
		t.Fatalf("expected forbidden error, got %s", appErr.Code)
	}
	if appErr.Message != "you do not have access to this task" {
		t.Fatalf("unexpected error message: %q", appErr.Message)
	}
}
