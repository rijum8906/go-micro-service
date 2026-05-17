package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type stubQuerier struct {
	db.Querier

	getActiveProjectMembershipFn func(context.Context, db.GetActiveProjectMembershipParams) (db.ProjectMembership, error)
	getTaskFn                    func(context.Context, uuid.UUID) (db.Task, error)
}

func (s stubQuerier) GetActiveProjectMembership(ctx context.Context, params db.GetActiveProjectMembershipParams) (db.ProjectMembership, error) {
	if s.getActiveProjectMembershipFn == nil {
		panic("unexpected GetActiveProjectMembership call")
	}
	return s.getActiveProjectMembershipFn(ctx, params)
}

func (s stubQuerier) GetTask(ctx context.Context, id uuid.UUID) (db.Task, error) {
	if s.getTaskFn == nil {
		panic("unexpected GetTask call")
	}
	return s.getTaskFn(ctx, id)
}

type stubTupleManager struct {
	checkFn func(context.Context, client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError)
}

func (stubTupleManager) Write(context.Context, []client.ClientTupleKey) *apperror.AppError {
	panic("unexpected Write call")
}

func (stubTupleManager) Read(context.Context, client.ClientReadRequest) (*client.ClientReadResponse, *apperror.AppError) {
	panic("unexpected Read call")
}

func (s stubTupleManager) Check(ctx context.Context, req client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
	if s.checkFn == nil {
		panic("unexpected Check call")
	}
	return s.checkFn(ctx, req)
}

func (stubTupleManager) Delete(context.Context, []client.ClientTupleKeyWithoutCondition) *apperror.AppError {
	panic("unexpected Delete call")
}

func TestRequireTaskRoleAllowsCreatorForNonProjectTask(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()

	authorizer := newAuthorizer(t, stubQuerier{
		getTaskFn: func(_ context.Context, id uuid.UUID) (db.Task, error) {
			return db.Task{ID: id, CreatedBy: userID}, nil
		},
	}, nil)

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
	creatorID := uuid.New()
	userID := uuid.New()

	authorizer := newAuthorizer(t, stubQuerier{
		getTaskFn: func(_ context.Context, id uuid.UUID) (db.Task, error) {
			return db.Task{ID: id, CreatedBy: creatorID}, nil
		},
	}, nil)

	task, appErr := authorizer.RequireTaskRole(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, RoleMember)
	if task != nil {
		t.Fatalf("expected nil task, got %#v", task)
	}
	if appErr == nil || appErr.Code != apperror.CodeForbidden {
		t.Fatalf("expected forbidden error, got %#v", appErr)
	}
}

func TestRequireProjectRoleUsesOpenFGACheck(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	allowed := true

	authorizer := newAuthorizer(t, stubQuerier{}, stubTupleManager{
		checkFn: func(_ context.Context, req client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
			if req.User != "user:"+userID.String() {
				t.Fatalf("unexpected user: %s", req.User)
			}
			if req.Relation != "can_manage_tasks" {
				t.Fatalf("unexpected relation: %s", req.Relation)
			}
			if req.Object != "project:"+projectID.String() {
				t.Fatalf("unexpected object: %s", req.Object)
			}
			return &client.ClientCheckResponse{
				CheckResponse: openfga.CheckResponse{Allowed: &allowed},
			}, nil
		},
	})

	membership, appErr := authorizer.RequireProjectRole(context.Background(), projectID, &dto.UserInfo{UserID: userID.String()}, RoleAdmin)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if membership == nil || membership.ProjectID != projectID || membership.UserID != userID || membership.Role != string(RoleAdmin) {
		t.Fatalf("unexpected membership: %#v", membership)
	}
}

func TestRequireProjectRoleRejectsDeniedOpenFGACheck(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	allowed := false

	authorizer := newAuthorizer(t, stubQuerier{}, stubTupleManager{
		checkFn: func(context.Context, client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
			return &client.ClientCheckResponse{
				CheckResponse: openfga.CheckResponse{Allowed: &allowed},
			}, nil
		},
	})

	membership, appErr := authorizer.RequireProjectRole(context.Background(), projectID, &dto.UserInfo{UserID: userID.String()}, RoleMember)
	if membership != nil {
		t.Fatalf("expected nil membership, got %#v", membership)
	}
	if appErr == nil || appErr.Code != apperror.CodeForbidden {
		t.Fatalf("expected forbidden error, got %#v", appErr)
	}
}

func TestRequireTaskRoleLoadsTaskAndUsesOpenFGACheck(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()
	allowed := true

	authorizer := newAuthorizer(t, stubQuerier{
		getTaskFn: func(_ context.Context, id uuid.UUID) (db.Task, error) {
			return db.Task{ID: id, CreatedBy: uuid.New()}, nil
		},
	}, stubTupleManager{
		checkFn: func(_ context.Context, req client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
			if req.User != "user:"+userID.String() {
				t.Fatalf("unexpected user: %s", req.User)
			}
			if req.Relation != "can_manage" {
				t.Fatalf("unexpected relation: %s", req.Relation)
			}
			if req.Object != "task:"+taskID.String() {
				t.Fatalf("unexpected object: %s", req.Object)
			}
			return &client.ClientCheckResponse{
				CheckResponse: openfga.CheckResponse{Allowed: &allowed},
			}, nil
		},
	})

	task, appErr := authorizer.RequireTaskRole(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, RoleAdmin)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if task == nil || task.ID != taskID {
		t.Fatalf("unexpected task: %#v", task)
	}
}

func newAuthorizer(t *testing.T, q db.Querier, tuples coreopenfga.TuppleManager) Authorizer {
	t.Helper()

	authorizer, appErr := NewAuthorizer(q, tuples)
	if appErr != nil {
		t.Fatalf("failed to construct authorizer: %v", appErr)
	}

	return authorizer
}

var _ db.Querier = stubQuerier{}
