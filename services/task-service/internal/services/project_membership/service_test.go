package projectmembership

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	servicetestutil "github.com/rijum8906/relay/services/task-service/internal/services/testutil"
)

type stubProjectMembershipRepository struct {
	addProjectMemberFn           func(context.Context, db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError)
	getActiveProjectMembershipFn func(context.Context, db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError)
	updateProjectMemberRoleFn    func(context.Context, db.UpdateProjectMemberRoleParams) (*db.ProjectMembership, *apperror.AppError)
	removeProjectMemberFn        func(context.Context, db.RemoveProjectMemberParams) (*db.ProjectMembership, *apperror.AppError)
	listProjectMembersFn         func(context.Context, uuid.UUID) ([]db.ProjectMembership, *apperror.AppError)
	listMembershipsByUserFn      func(context.Context, uuid.UUID) ([]db.ProjectMembership, *apperror.AppError)
}

type stubProjectMembershipAuthorizer struct {
	requireProjectRoleFn func(context.Context, uuid.UUID, *dto.UserInfo, authz.Role) (*db.ProjectMembership, *apperror.AppError)
}

func (s stubProjectMembershipAuthorizer) RequireProjectRole(ctx context.Context, projectID uuid.UUID, userInfo *dto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
	if s.requireProjectRoleFn == nil {
		panic("unexpected RequireProjectRole call")
	}
	return s.requireProjectRoleFn(ctx, projectID, userInfo, minRole)
}

func (stubProjectMembershipAuthorizer) RequireTaskRole(context.Context, uuid.UUID, *dto.UserInfo, authz.Role) (*db.Task, *apperror.AppError) {
	panic("unexpected RequireTaskRole call")
}

func (s *stubProjectMembershipRepository) AddProjectMember(ctx context.Context, params db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
	if s.addProjectMemberFn == nil {
		panic("unexpected AddProjectMember call")
	}
	return s.addProjectMemberFn(ctx, params)
}

func (s *stubProjectMembershipRepository) GetActiveProjectMembership(ctx context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
	if s.getActiveProjectMembershipFn == nil {
		panic("unexpected GetActiveProjectMembership call")
	}
	return s.getActiveProjectMembershipFn(ctx, params)
}

func (s *stubProjectMembershipRepository) UpdateProjectMemberRole(ctx context.Context, params db.UpdateProjectMemberRoleParams) (*db.ProjectMembership, *apperror.AppError) {
	if s.updateProjectMemberRoleFn == nil {
		panic("unexpected UpdateProjectMemberRole call")
	}
	return s.updateProjectMemberRoleFn(ctx, params)
}

func (s *stubProjectMembershipRepository) RemoveProjectMember(ctx context.Context, params db.RemoveProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
	if s.removeProjectMemberFn == nil {
		panic("unexpected RemoveProjectMember call")
	}
	return s.removeProjectMemberFn(ctx, params)
}

func (s *stubProjectMembershipRepository) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
	if s.listProjectMembersFn == nil {
		panic("unexpected ListProjectMembers call")
	}
	return s.listProjectMembersFn(ctx, projectID)
}

func (s *stubProjectMembershipRepository) ListProjectMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
	if s.listMembershipsByUserFn == nil {
		panic("unexpected ListProjectMembershipsByUser call")
	}
	return s.listMembershipsByUserFn(ctx, userID)
}

func TestNewProjectMembershipService(t *testing.T) {
	svc, err := NewProjectMembershipService(nil, servicetestutil.NewAllowAuthorizer())
	if err == nil {
		t.Fatal("expected constructor error for nil repository")
	}
	if svc != nil {
		t.Fatal("expected nil service when repository is nil")
	}
	if err.Code != apperror.CodeInternal {
		t.Fatalf("expected internal error, got %s", err.Code)
	}

	svc, err = NewProjectMembershipService(&stubProjectMembershipRepository{}, servicetestutil.NewAllowAuthorizer())
	if err != nil {
		t.Fatalf("expected constructor success, got error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestAddProjectMemberValidation(t *testing.T) {
	testCases := []struct {
		name        string
		req         *taskv1.AddProjectMemberRequest
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			wantMessage: "add project member request is required",
		},
		{
			name:        "missing project id",
			req:         &taskv1.AddProjectMemberRequest{UserId: uuid.NewString()},
			wantMessage: "project id is required",
		},
		{
			name:        "missing user id",
			req:         &taskv1.AddProjectMemberRequest{ProjectId: uuid.NewString()},
			wantMessage: "user id is required",
		},
		{
			name:        "invalid role",
			req:         &taskv1.AddProjectMemberRequest{ProjectId: uuid.NewString(), UserId: uuid.NewString(), Role: "viewer"},
			wantMessage: "invalid project member role",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
				getActiveProjectMembershipFn: func(context.Context, db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
				addProjectMemberFn: func(context.Context, db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.AddProjectMember(context.Background(), tc.req, &dto.UserInfo{UserID: uuid.NewString()})
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertProjectMembershipAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestAddProjectMemberDuplicateConflict(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		getActiveProjectMembershipFn: func(_ context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID {
				t.Fatalf("unexpected project id: %s", params.ProjectID)
			}
			if params.UserID != userID {
				t.Fatalf("unexpected user id: %s", params.UserID)
			}
			return &db.ProjectMembership{ID: uuid.New(), ProjectID: projectID, UserID: userID, Role: "member"}, nil
		},
	})

	res, err := svc.AddProjectMember(context.Background(), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "member",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertProjectMembershipAppError(t, err, apperror.CodeConflict, "project member already exists")
}

func TestAddProjectMemberLookupErrorPropagation(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to get active membership"}

	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		getActiveProjectMembershipFn: func(_ context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID || params.UserID != userID {
				t.Fatalf("unexpected lookup params: %#v", params)
			}
			return nil, repoErr
		},
	})

	res, err := svc.AddProjectMember(context.Background(), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "member",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestAddProjectMemberSuccess(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	membershipID := uuid.New()
	joinedAt := pgtype.Timestamptz{Time: time.Date(2026, time.May, 4, 9, 0, 0, 0, time.UTC), Valid: true}

	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		getActiveProjectMembershipFn: func(_ context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID || params.UserID != userID {
				t.Fatalf("unexpected lookup params: %#v", params)
			}
			return nil, &apperror.AppError{Code: apperror.CodeNotFound, Message: "project membership not found"}
		},
		addProjectMemberFn: func(_ context.Context, params db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID {
				t.Fatalf("unexpected project id: %s", params.ProjectID)
			}
			if params.UserID != userID {
				t.Fatalf("unexpected user id: %s", params.UserID)
			}
			if params.Role != "admin" {
				t.Fatalf("unexpected role: %s", params.Role)
			}
			return &db.ProjectMembership{
				ID:        membershipID,
				ProjectID: projectID,
				UserID:    userID,
				Role:      params.Role,
				JoinedAt:  joinedAt,
			}, nil
		},
	})

	res, err := svc.AddProjectMember(context.Background(), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "admin",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != membershipID.String() {
		t.Fatalf("unexpected id: %s", res.Id)
	}
	if res.ProjectId != projectID.String() {
		t.Fatalf("unexpected project id: %s", res.ProjectId)
	}
	if res.UserId != userID.String() {
		t.Fatalf("unexpected user id: %s", res.UserId)
	}
	if res.Role != "admin" {
		t.Fatalf("unexpected role: %s", res.Role)
	}
}

func TestAddProjectMemberRequiresOwnerToAddOwner(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	repoCalled := false

	svc := mustProjectMembershipServiceWithAuthorizer(t, &stubProjectMembershipRepository{
		getActiveProjectMembershipFn: func(context.Context, db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
			repoCalled = true
			return nil, nil
		},
		addProjectMemberFn: func(context.Context, db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
			repoCalled = true
			return nil, nil
		},
	}, stubProjectMembershipAuthorizer{
		requireProjectRoleFn: func(_ context.Context, gotProjectID uuid.UUID, gotUserInfo *dto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
			if gotProjectID != projectID {
				t.Fatalf("unexpected project id: %s", gotProjectID)
			}
			if gotUserInfo == nil || gotUserInfo.UserID == "" {
				t.Fatalf("unexpected user info: %#v", gotUserInfo)
			}
			if minRole != authz.RoleOwner {
				t.Fatalf("expected owner role requirement, got %s", minRole)
			}
			return nil, apperror.ErrForbidden.WithMessage("insufficient project role")
		},
	})

	res, err := svc.AddProjectMember(context.Background(), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "owner",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertProjectMembershipAppError(t, err, apperror.CodeForbidden, "insufficient project role")
	if repoCalled {
		t.Fatal("repository should not be called when owner authorization fails")
	}
}

func TestAddProjectMemberOwnerCanAddOwner(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	svc := mustProjectMembershipServiceWithAuthorizer(t, &stubProjectMembershipRepository{
		getActiveProjectMembershipFn: func(_ context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID || params.UserID != userID {
				t.Fatalf("unexpected lookup params: %#v", params)
			}
			return nil, &apperror.AppError{Code: apperror.CodeNotFound, Message: "project membership not found"}
		},
		addProjectMemberFn: func(_ context.Context, params db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.Role != "owner" {
				t.Fatalf("unexpected role: %s", params.Role)
			}
			return &db.ProjectMembership{
				ID:        uuid.New(),
				ProjectID: params.ProjectID,
				UserID:    params.UserID,
				Role:      params.Role,
			}, nil
		},
	}, stubProjectMembershipAuthorizer{
		requireProjectRoleFn: func(_ context.Context, gotProjectID uuid.UUID, gotUserInfo *dto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
			if gotProjectID != projectID {
				t.Fatalf("unexpected project id: %s", gotProjectID)
			}
			if gotUserInfo == nil || gotUserInfo.UserID == "" {
				t.Fatalf("unexpected user info: %#v", gotUserInfo)
			}
			if minRole != authz.RoleOwner {
				t.Fatalf("expected owner role requirement, got %s", minRole)
			}
			return &db.ProjectMembership{ProjectID: projectID, Role: string(minRole)}, nil
		},
	})

	res, err := svc.AddProjectMember(context.Background(), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "owner",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil || res.Role != "owner" {
		t.Fatalf("unexpected response: %#v", res)
	}
}

func TestAddProjectMemberAdminCanAddMember(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	svc := mustProjectMembershipServiceWithAuthorizer(t, &stubProjectMembershipRepository{
		getActiveProjectMembershipFn: func(_ context.Context, params db.GetActiveProjectMembershipParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID || params.UserID != userID {
				t.Fatalf("unexpected lookup params: %#v", params)
			}
			return nil, &apperror.AppError{Code: apperror.CodeNotFound, Message: "project membership not found"}
		},
		addProjectMemberFn: func(_ context.Context, params db.AddProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.Role != "member" {
				t.Fatalf("unexpected role: %s", params.Role)
			}
			return &db.ProjectMembership{
				ID:        uuid.New(),
				ProjectID: params.ProjectID,
				UserID:    params.UserID,
				Role:      params.Role,
			}, nil
		},
	}, stubProjectMembershipAuthorizer{
		requireProjectRoleFn: func(_ context.Context, gotProjectID uuid.UUID, gotUserInfo *dto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
			if gotProjectID != projectID {
				t.Fatalf("unexpected project id: %s", gotProjectID)
			}
			if gotUserInfo == nil || gotUserInfo.UserID == "" {
				t.Fatalf("unexpected user info: %#v", gotUserInfo)
			}
			if minRole != authz.RoleAdmin {
				t.Fatalf("expected admin role requirement, got %s", minRole)
			}
			return &db.ProjectMembership{ProjectID: projectID, Role: string(minRole)}, nil
		},
	})

	res, err := svc.AddProjectMember(context.Background(), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "member",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil || res.Role != "member" {
		t.Fatalf("unexpected response: %#v", res)
	}
}

func TestRemoveProjectMemberSuccess(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		removeProjectMemberFn: func(_ context.Context, params db.RemoveProjectMemberParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID {
				t.Fatalf("unexpected project id: %s", params.ProjectID)
			}
			if params.UserID != userID {
				t.Fatalf("unexpected user id: %s", params.UserID)
			}
			return &db.ProjectMembership{ID: uuid.New(), ProjectID: projectID, UserID: userID}, nil
		},
	})

	res, err := svc.RemoveProjectMember(context.Background(), &taskv1.RemoveProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("expected success response, got %#v", res)
	}
}

func TestUpdateProjectMemberRoleSuccess(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		updateProjectMemberRoleFn: func(_ context.Context, params db.UpdateProjectMemberRoleParams) (*db.ProjectMembership, *apperror.AppError) {
			if params.ProjectID != projectID {
				t.Fatalf("unexpected project id: %s", params.ProjectID)
			}
			if params.UserID != userID {
				t.Fatalf("unexpected user id: %s", params.UserID)
			}
			if params.Role != "owner" {
				t.Fatalf("unexpected role: %s", params.Role)
			}
			return &db.ProjectMembership{
				ID:        uuid.New(),
				ProjectID: projectID,
				UserID:    userID,
				Role:      params.Role,
			}, nil
		},
	})

	res, err := svc.UpdateProjectMemberRole(context.Background(), &taskv1.UpdateProjectMemberRoleRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      "owner",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.ProjectId != projectID.String() {
		t.Fatalf("unexpected project id: %s", res.ProjectId)
	}
	if res.UserId != userID.String() {
		t.Fatalf("unexpected user id: %s", res.UserId)
	}
	if res.Role != "owner" {
		t.Fatalf("unexpected role: %s", res.Role)
	}
}

func TestListProjectMembersSuccess(t *testing.T) {
	projectID := uuid.New()
	memberID := uuid.New()
	userID := uuid.New()

	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		listProjectMembersFn: func(_ context.Context, got uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
			if got != projectID {
				t.Fatalf("unexpected project id: %s", got)
			}
			return []db.ProjectMembership{
				{
					ID:        memberID,
					ProjectID: projectID,
					UserID:    userID,
					Role:      "member",
				},
			}, nil
		},
	})

	res, err := svc.ListProjectMembers(context.Background(), &taskv1.ListProjectMembersRequest{
		ProjectId: projectID.String(),
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(res.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(res.Members))
	}
	if res.Members[0].Id != memberID.String() {
		t.Fatalf("unexpected membership id: %s", res.Members[0].Id)
	}
}

func TestListProjectMembersValidationAndRepoError(t *testing.T) {
	svc := mustProjectMembershipService(t, &stubProjectMembershipRepository{
		listProjectMembersFn: func(context.Context, uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid project id")
			return nil, nil
		},
	})

	res, err := svc.ListProjectMembers(context.Background(), &taskv1.ListProjectMembersRequest{
		ProjectId: "bad-uuid",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertProjectMembershipAppError(t, err, apperror.CodeValidation, "invalid uuid")

	projectID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to list project members"}
	svc = mustProjectMembershipService(t, &stubProjectMembershipRepository{
		listProjectMembersFn: func(_ context.Context, got uuid.UUID) ([]db.ProjectMembership, *apperror.AppError) {
			if got != projectID {
				t.Fatalf("unexpected project id: %s", got)
			}
			return nil, repoErr
		},
	})

	res, err = svc.ListProjectMembers(context.Background(), &taskv1.ListProjectMembersRequest{
		ProjectId: projectID.String(),
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func mustProjectMembershipService(t *testing.T, repo *stubProjectMembershipRepository) ProjectMembershipService {
	t.Helper()

	return mustProjectMembershipServiceWithAuthorizer(t, repo, servicetestutil.NewAllowAuthorizer())
}

func mustProjectMembershipServiceWithAuthorizer(t *testing.T, repo *stubProjectMembershipRepository, authorizer authz.Authorizer) ProjectMembershipService {
	t.Helper()

	svc, err := NewProjectMembershipService(repo, authorizer)
	if err != nil {
		t.Fatalf("failed to construct project membership service: %v", err)
	}

	return svc
}

func assertProjectMembershipAppError(t *testing.T, err *apperror.AppError, code apperror.ErrorCode, message string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected app error, got nil")
	}
	if err.Code != code {
		t.Fatalf("expected error code %s, got %s", code, err.Code)
	}
	if err.Message != message {
		t.Fatalf("expected error message %q, got %q", message, err.Message)
	}
}
