package project

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	servicetestutil "github.com/rijum8906/relay/services/task-service/internal/services/testutil"
)

type stubProjectRepository struct {
	createProjectFn   func(context.Context, db.CreateProjectParams) (*db.Project, *apperror.AppError)
	getProjectFn      func(context.Context, uuid.UUID) (*db.Project, *apperror.AppError)
	updateProjectFn   func(context.Context, db.UpdateProjectParams) (*db.Project, *apperror.AppError)
	completeProjectFn func(context.Context, uuid.UUID) (*db.Project, *apperror.AppError)
	archiveProjectFn  func(context.Context, uuid.UUID) (*db.Project, *apperror.AppError)
	deleteProjectFn   func(context.Context, db.DeleteProjectParams) (*db.Project, *apperror.AppError)
	listProjectsFn    func(context.Context, db.ListProjectsParams) ([]db.Project, *apperror.AppError)
}

func (s *stubProjectRepository) CreateProject(ctx context.Context, params db.CreateProjectParams) (*db.Project, *apperror.AppError) {
	if s.createProjectFn == nil {
		panic("unexpected CreateProject call")
	}
	return s.createProjectFn(ctx, params)
}

func (s *stubProjectRepository) GetProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
	if s.getProjectFn == nil {
		panic("unexpected GetProject call")
	}
	return s.getProjectFn(ctx, id)
}

func (s *stubProjectRepository) UpdateProject(ctx context.Context, params db.UpdateProjectParams) (*db.Project, *apperror.AppError) {
	if s.updateProjectFn == nil {
		panic("unexpected UpdateProject call")
	}
	return s.updateProjectFn(ctx, params)
}

func (s *stubProjectRepository) CompleteProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
	if s.completeProjectFn == nil {
		panic("unexpected CompleteProject call")
	}
	return s.completeProjectFn(ctx, id)
}

func (s *stubProjectRepository) ArchiveProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
	if s.archiveProjectFn == nil {
		panic("unexpected ArchiveProject call")
	}
	return s.archiveProjectFn(ctx, id)
}

func (s *stubProjectRepository) DeleteProject(ctx context.Context, params db.DeleteProjectParams) (*db.Project, *apperror.AppError) {
	if s.deleteProjectFn == nil {
		panic("unexpected DeleteProject call")
	}
	return s.deleteProjectFn(ctx, params)
}

func (s *stubProjectRepository) ListProjects(ctx context.Context, params db.ListProjectsParams) ([]db.Project, *apperror.AppError) {
	if s.listProjectsFn == nil {
		panic("unexpected ListProjects call")
	}
	return s.listProjectsFn(ctx, params)
}

func TestNewProjectService(t *testing.T) {
	svc, err := NewProjectService(nil, servicetestutil.NewAllowAuthorizer())
	if err == nil {
		t.Fatal("expected constructor error for nil repository")
	}
	if svc != nil {
		t.Fatal("expected nil service when repository is nil")
	}
	if err.Code != apperror.CodeInternal {
		t.Fatalf("expected internal error, got %s", err.Code)
	}

	svc, err = NewProjectService(&stubProjectRepository{}, servicetestutil.NewAllowAuthorizer())
	if err != nil {
		t.Fatalf("expected constructor success, got error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestCreateProjectValidation(t *testing.T) {
	validUser := &coredto.UserInfo{UserID: uuid.NewString()}

	testCases := []struct {
		name        string
		req         *taskv1.CreateProjectRequest
		userInfo    *coredto.UserInfo
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			userInfo:    validUser,
			wantMessage: "create project request is required",
		},
		{
			name:        "missing user metadata",
			req:         &taskv1.CreateProjectRequest{Name: "Platform"},
			userInfo:    nil,
			wantMessage: "user metadata is required",
		},
		{
			name:        "missing name",
			req:         &taskv1.CreateProjectRequest{},
			userInfo:    validUser,
			wantMessage: "name is required",
		},
		{
			name:        "invalid organization id",
			req:         &taskv1.CreateProjectRequest{Name: "Platform", OrganizationId: "bad-uuid"},
			userInfo:    validUser,
			wantMessage: "invalid uuid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustProjectService(t, &stubProjectRepository{
				createProjectFn: func(context.Context, db.CreateProjectParams) (*db.Project, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.CreateProject(context.Background(), tc.req, tc.userInfo)
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertProjectAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestCreateProjectSuccess(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	projectID := uuid.New()
	now := pgtype.Timestamptz{Time: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC), Valid: true}

	svc := mustProjectService(t, &stubProjectRepository{
		createProjectFn: func(_ context.Context, params db.CreateProjectParams) (*db.Project, *apperror.AppError) {
			if !params.OrganizationID.Valid || params.OrganizationID.Bytes != orgID {
				t.Fatalf("unexpected organization id: %#v", params.OrganizationID)
			}
			if params.CreatedBy != userID {
				t.Fatalf("unexpected created_by: %s", params.CreatedBy)
			}
			if params.Name != "Platform" {
				t.Fatalf("unexpected name: %s", params.Name)
			}
			if params.Description != "Core workstream" {
				t.Fatalf("unexpected description: %s", params.Description)
			}

			return &db.Project{
				ID:             projectID,
				OrganizationID: params.OrganizationID,
				CreatedBy:      params.CreatedBy,
				Name:           params.Name,
				Description:    params.Description,
				Status:         "active",
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
	})

	res, err := svc.CreateProject(context.Background(), &taskv1.CreateProjectRequest{
		OrganizationId: orgID.String(),
		Name:           "Platform",
		Description:    "Core workstream",
	}, &coredto.UserInfo{UserID: userID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != projectID.String() {
		t.Fatalf("unexpected id: %s", res.Id)
	}
	if res.OrganizationId != orgID.String() {
		t.Fatalf("unexpected organization id: %s", res.OrganizationId)
	}
	if res.CreatedBy != userID.String() {
		t.Fatalf("unexpected created_by: %s", res.CreatedBy)
	}
	if res.Name != "Platform" {
		t.Fatalf("unexpected name: %s", res.Name)
	}
}

func TestGetProjectValidation(t *testing.T) {
	testCases := []struct {
		name        string
		req         *taskv1.GetProjectRequest
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			wantMessage: "get project request is required",
		},
		{
			name:        "missing id",
			req:         &taskv1.GetProjectRequest{},
			wantMessage: "project id is required",
		},
		{
			name:        "invalid id",
			req:         &taskv1.GetProjectRequest{Id: "bad-uuid"},
			wantMessage: "invalid uuid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustProjectService(t, &stubProjectRepository{
				getProjectFn: func(context.Context, uuid.UUID) (*db.Project, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.GetProject(context.Background(), tc.req, &dto.UserInfo{UserID: uuid.NewString()})
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertProjectAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestGetProjectRepoError(t *testing.T) {
	projectID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeNotFound, Message: "project not found"}

	svc := mustProjectService(t, &stubProjectRepository{
		getProjectFn: func(_ context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
			if id != projectID {
				t.Fatalf("unexpected project id: %s", id)
			}
			return nil, repoErr
		},
	})

	res, err := svc.GetProject(context.Background(), &taskv1.GetProjectRequest{Id: projectID.String()}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestUpdateProjectSuccess(t *testing.T) {
	projectID := uuid.New()

	svc := mustProjectService(t, &stubProjectRepository{
		updateProjectFn: func(_ context.Context, params db.UpdateProjectParams) (*db.Project, *apperror.AppError) {
			if params.ID != projectID {
				t.Fatalf("unexpected project id: %s", params.ID)
			}
			if params.Name != "Renamed" {
				t.Fatalf("unexpected name: %s", params.Name)
			}
			if params.Description != "Updated description" {
				t.Fatalf("unexpected description: %s", params.Description)
			}

			return &db.Project{
				ID:          params.ID,
				CreatedBy:   uuid.New(),
				Name:        params.Name,
				Description: params.Description,
				Status:      "active",
			}, nil
		},
	})

	res, err := svc.UpdateProject(context.Background(), &taskv1.UpdateProjectRequest{
		Id:          projectID.String(),
		Name:        "Renamed",
		Description: "Updated description",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != projectID.String() {
		t.Fatalf("unexpected id: %s", res.Id)
	}
	if res.Name != "Renamed" {
		t.Fatalf("unexpected name: %s", res.Name)
	}
}

func TestArchiveProjectValidationAndRepoError(t *testing.T) {
	svc := mustProjectService(t, &stubProjectRepository{
		archiveProjectFn: func(context.Context, uuid.UUID) (*db.Project, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid archive request")
			return nil, nil
		},
	})

	res, err := svc.ArchiveProject(context.Background(), &taskv1.ArchiveProjectRequest{Id: "bad-uuid"}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertProjectAppError(t, err, apperror.CodeValidation, "invalid uuid")

	projectID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeNotFound, Message: "project not found"}
	svc = mustProjectService(t, &stubProjectRepository{
		archiveProjectFn: func(_ context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
			if id != projectID {
				t.Fatalf("unexpected project id: %s", id)
			}
			return nil, repoErr
		},
	})

	res, err = svc.ArchiveProject(context.Background(), &taskv1.ArchiveProjectRequest{Id: projectID.String()}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestDeleteProjectValidation(t *testing.T) {
	validUser := &coredto.UserInfo{UserID: uuid.NewString()}

	testCases := []struct {
		name        string
		req         *taskv1.DeleteProjectRequest
		userInfo    *coredto.UserInfo
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			userInfo:    validUser,
			wantMessage: "delete project request is required",
		},
		{
			name:        "missing user metadata",
			req:         &taskv1.DeleteProjectRequest{Id: uuid.NewString()},
			userInfo:    nil,
			wantMessage: "user metadata is required",
		},
		{
			name:        "missing id",
			req:         &taskv1.DeleteProjectRequest{},
			userInfo:    validUser,
			wantMessage: "project id is required",
		},
		{
			name:        "invalid id",
			req:         &taskv1.DeleteProjectRequest{Id: "bad-uuid"},
			userInfo:    validUser,
			wantMessage: "invalid uuid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustProjectService(t, &stubProjectRepository{
				deleteProjectFn: func(context.Context, db.DeleteProjectParams) (*db.Project, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.DeleteProject(context.Background(), tc.req, tc.userInfo)
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertProjectAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestDeleteProjectSuccess(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	svc := mustProjectService(t, &stubProjectRepository{
		deleteProjectFn: func(_ context.Context, params db.DeleteProjectParams) (*db.Project, *apperror.AppError) {
			if params.ID != projectID {
				t.Fatalf("unexpected project id: %s", params.ID)
			}
			if !params.DeletedBy.Valid || params.DeletedBy.Bytes != userID {
				t.Fatalf("unexpected deleted_by: %#v", params.DeletedBy)
			}
			return &db.Project{ID: projectID}, nil
		},
	})

	res, err := svc.DeleteProject(context.Background(), &taskv1.DeleteProjectRequest{
		Id: projectID.String(),
	}, &coredto.UserInfo{UserID: userID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("expected success response, got %#v", res)
	}
}

func TestListProjectsSuccess(t *testing.T) {
	orgID := uuid.New()
	projectID := uuid.New()

	svc := mustProjectService(t, &stubProjectRepository{
		listProjectsFn: func(_ context.Context, params db.ListProjectsParams) ([]db.Project, *apperror.AppError) {
			if !params.OrganizationID.Valid || params.OrganizationID.Bytes != orgID {
				t.Fatalf("unexpected organization id: %#v", params.OrganizationID)
			}
			if params.Status != "active" {
				t.Fatalf("unexpected status: %s", params.Status)
			}
			return []db.Project{
				{
					ID:             projectID,
					OrganizationID: params.OrganizationID,
					CreatedBy:      uuid.New(),
					Name:           "Platform",
					Description:    "Core workstream",
					Status:         "active",
				},
			}, nil
		},
	})

	res, err := svc.ListProjects(context.Background(), &taskv1.ListProjectsRequest{
		OrganizationId: orgID.String(),
		Status:         "active",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(res.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(res.Projects))
	}
	if res.Projects[0].Id != projectID.String() {
		t.Fatalf("unexpected id: %s", res.Projects[0].Id)
	}
}

func TestListProjectsInvalidOrganizationAndRepoError(t *testing.T) {
	svc := mustProjectService(t, &stubProjectRepository{
		listProjectsFn: func(context.Context, db.ListProjectsParams) ([]db.Project, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid organization id")
			return nil, nil
		},
	})

	res, err := svc.ListProjects(context.Background(), &taskv1.ListProjectsRequest{
		OrganizationId: "bad-uuid",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertProjectAppError(t, err, apperror.CodeValidation, "invalid uuid")

	repoErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to list projects"}
	svc = mustProjectService(t, &stubProjectRepository{
		listProjectsFn: func(_ context.Context, params db.ListProjectsParams) ([]db.Project, *apperror.AppError) {
			if params.Status != "archived" {
				t.Fatalf("unexpected status: %s", params.Status)
			}
			return nil, repoErr
		},
	})

	res, err = svc.ListProjects(context.Background(), &taskv1.ListProjectsRequest{Status: "archived"}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestCompleteProjectSuccess(t *testing.T) {
	projectID := uuid.New()
	completedAt := pgtype.Timestamptz{Time: time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC), Valid: true}

	svc := mustProjectService(t, &stubProjectRepository{
		completeProjectFn: func(_ context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
			if id != projectID {
				t.Fatalf("unexpected project id: %s", id)
			}
			return &db.Project{
				ID:          id,
				CreatedBy:   uuid.New(),
				Name:        "Platform",
				Status:      "completed",
				CompletedAt: completedAt,
			}, nil
		},
	})

	res, err := svc.CompleteProject(context.Background(), &taskv1.CompleteProjectRequest{Id: projectID.String()}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Status != "completed" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
	if res.CompletedAt == nil || !res.CompletedAt.AsTime().Equal(completedAt.Time) {
		t.Fatalf("unexpected completed_at: %#v", res.CompletedAt)
	}
}

func mustProjectService(t *testing.T, repo *stubProjectRepository) ProjectService {
	t.Helper()

	svc, err := NewProjectService(repo, servicetestutil.NewAllowAuthorizer())
	if err != nil {
		t.Fatalf("failed to construct project service: %v", err)
	}

	return svc
}

func assertProjectAppError(t *testing.T, err *apperror.AppError, code apperror.ErrorCode, message string) {
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
