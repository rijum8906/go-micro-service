package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/dto"
	taskpermissions "github.com/rijum8906/relay/packages/core/permissions/task"
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

func TestRequireTaskPermissionAllowsCreatorForNonProjectTask(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()

	authorizer := newAuthorizer(t, stubQuerier{
		getTaskFn: func(_ context.Context, id uuid.UUID) (db.Task, error) {
			return db.Task{ID: id, CreatedBy: userID}, nil
		},
	}, nil)

	task, appErr := authorizer.RequireTaskPermission(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, taskpermissions.PermissionCanDelete)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if task == nil || task.ID != taskID {
		t.Fatalf("unexpected task: %#v", task)
	}
}

func TestRequireTaskPermissionRejectsNonCreatorForNonProjectTask(t *testing.T) {
	taskID := uuid.New()
	creatorID := uuid.New()
	userID := uuid.New()

	authorizer := newAuthorizer(t, stubQuerier{
		getTaskFn: func(_ context.Context, id uuid.UUID) (db.Task, error) {
			return db.Task{ID: id, CreatedBy: creatorID}, nil
		},
	}, nil)

	task, appErr := authorizer.RequireTaskPermission(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, taskpermissions.PermissionCanView)
	if task != nil {
		t.Fatalf("expected nil task, got %#v", task)
	}
	if appErr == nil || appErr.Code != apperror.CodeForbidden {
		t.Fatalf("expected forbidden error, got %#v", appErr)
	}
}

func TestRequireProjectPermissionUsesOpenFGACheck(t *testing.T) {
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

	membership, appErr := authorizer.RequireProjectPermission(context.Background(), projectID, &dto.UserInfo{UserID: userID.String()}, taskpermissions.PermissionCanManageTasks)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if membership == nil || membership.ProjectID != projectID || membership.UserID != userID {
		t.Fatalf("unexpected membership: %#v", membership)
	}
}

func TestRequireProjectPermissionRejectsDeniedOpenFGACheck(t *testing.T) {
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

	membership, appErr := authorizer.RequireProjectPermission(context.Background(), projectID, &dto.UserInfo{UserID: userID.String()}, taskpermissions.PermissionCanView)
	if membership != nil {
		t.Fatalf("expected nil membership, got %#v", membership)
	}
	if appErr == nil || appErr.Code != apperror.CodeForbidden {
		t.Fatalf("expected forbidden error, got %#v", appErr)
	}
}

func TestRequireTaskPermissionLoadsTaskAndUsesOpenFGACheck(t *testing.T) {
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
			if req.Relation != taskpermissions.PermissionCanAssign {
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

	task, appErr := authorizer.RequireTaskPermission(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, taskpermissions.PermissionCanAssign)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if task == nil || task.ID != taskID {
		t.Fatalf("unexpected task: %#v", task)
	}
}

func TestRequireTaskPermissionUsesProjectScopedCustomPermissionObject(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	allowed := true
	checks := 0

	authorizer := newAuthorizer(t, stubQuerier{
		getTaskFn: func(_ context.Context, id uuid.UUID) (db.Task, error) {
			return db.Task{
				ID:        id,
				CreatedBy: uuid.New(),
				ProjectID: pgtype.UUID{
					Bytes: projectID,
					Valid: true,
				},
			}, nil
		},
	}, stubTupleManager{
		checkFn: func(_ context.Context, req client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
			checks++
			if checks == 1 {
				if req.Relation != taskpermissions.PermissionCanComment || req.Object != "task:"+taskID.String() {
					t.Fatalf("unexpected direct check: %#v", req)
				}
				denied := false
				return &client.ClientCheckResponse{
					CheckResponse: openfga.CheckResponse{Allowed: &denied},
				}, nil
			}
			if req.User != "user:"+userID.String() {
				t.Fatalf("unexpected user: %s", req.User)
			}
			if req.Relation != "allowed" {
				t.Fatalf("unexpected relation: %s", req.Relation)
			}
			expectedObject := taskpermissions.GeneratePermissionObject(projectID.String(), taskpermissions.ResourceTask, taskpermissions.PermissionCanComment)
			if req.Object != expectedObject {
				t.Fatalf("unexpected object: %s", req.Object)
			}
			return &client.ClientCheckResponse{
				CheckResponse: openfga.CheckResponse{Allowed: &allowed},
			}, nil
		},
	})

	task, appErr := authorizer.RequireTaskPermission(context.Background(), taskID, &dto.UserInfo{UserID: userID.String()}, taskpermissions.PermissionCanComment)
	if appErr != nil {
		t.Fatalf("expected success, got error: %v", appErr)
	}
	if task == nil || task.ID != taskID {
		t.Fatalf("unexpected task: %#v", task)
	}
	if checks != 2 {
		t.Fatalf("expected two FGA checks, got %d", checks)
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
