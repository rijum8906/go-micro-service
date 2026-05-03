package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	servicetestutil "github.com/rijum8906/relay/services/task-service/internal/services/testutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stubTaskRepository struct {
	createTaskFn              func(context.Context, db.CreateTaskParams) (*db.Task, *apperror.AppError)
	getTaskFn                 func(context.Context, uuid.UUID) (*db.Task, *apperror.AppError)
	listTasksByProjectFn      func(context.Context, pgtype.UUID) ([]db.Task, *apperror.AppError)
	updateTaskFn              func(context.Context, db.UpdateTaskParams) (*db.Task, *apperror.AppError)
	deleteTaskFn              func(context.Context, db.DeleteTaskParams) (*db.Task, *apperror.AppError)
	archiveTaskFn             func(context.Context, db.ArchiveTaskParams) (*db.Task, *apperror.AppError)
	updateTaskStatusFn        func(context.Context, db.UpdateTaskStatusParams) (*db.Task, *apperror.AppError)
	updateTaskProgressFn      func(context.Context, db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError)
	listTasksByOrganizationFn func(context.Context, db.ListTasksByOrganizationParams) ([]db.Task, *apperror.AppError)
	listTasksByParentFn       func(context.Context, pgtype.UUID) ([]db.Task, *apperror.AppError)
	listTasksByCreatorFn      func(context.Context, db.ListTasksByCreatorParams) ([]db.Task, *apperror.AppError)
}

type stubTaskAuthorizer struct {
	requireProjectRoleFn func(context.Context, uuid.UUID, *dto.UserInfo, authz.Role) (*db.ProjectMembership, *apperror.AppError)
	requireTaskRoleFn    func(context.Context, uuid.UUID, *dto.UserInfo, authz.Role) (*db.Task, *apperror.AppError)
}

func (s stubTaskAuthorizer) RequireProjectRole(ctx context.Context, projectID uuid.UUID, userInfo *dto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
	if s.requireProjectRoleFn == nil {
		panic("unexpected RequireProjectRole call")
	}
	return s.requireProjectRoleFn(ctx, projectID, userInfo, minRole)
}

func (s stubTaskAuthorizer) RequireTaskRole(ctx context.Context, taskID uuid.UUID, userInfo *dto.UserInfo, minRole authz.Role) (*db.Task, *apperror.AppError) {
	if s.requireTaskRoleFn == nil {
		panic("unexpected RequireTaskRole call")
	}
	return s.requireTaskRoleFn(ctx, taskID, userInfo, minRole)
}

func (s *stubTaskRepository) CreateTask(ctx context.Context, params db.CreateTaskParams) (*db.Task, *apperror.AppError) {
	if s.createTaskFn == nil {
		panic("unexpected CreateTask call")
	}
	return s.createTaskFn(ctx, params)
}

func (s *stubTaskRepository) GetTask(ctx context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
	if s.getTaskFn == nil {
		panic("unexpected GetTask call")
	}
	return s.getTaskFn(ctx, id)
}

func (s *stubTaskRepository) ListTasksByProject(ctx context.Context, projectID pgtype.UUID) ([]db.Task, *apperror.AppError) {
	if s.listTasksByProjectFn == nil {
		panic("unexpected ListTasksByProject call")
	}
	return s.listTasksByProjectFn(ctx, projectID)
}

func (s *stubTaskRepository) UpdateTask(ctx context.Context, params db.UpdateTaskParams) (*db.Task, *apperror.AppError) {
	if s.updateTaskFn == nil {
		panic("unexpected UpdateTask call")
	}
	return s.updateTaskFn(ctx, params)
}

func (s *stubTaskRepository) DeleteTask(ctx context.Context, params db.DeleteTaskParams) (*db.Task, *apperror.AppError) {
	if s.deleteTaskFn == nil {
		panic("unexpected DeleteTask call")
	}
	return s.deleteTaskFn(ctx, params)
}

func (s *stubTaskRepository) ArchiveTask(ctx context.Context, params db.ArchiveTaskParams) (*db.Task, *apperror.AppError) {
	if s.archiveTaskFn == nil {
		panic("unexpected ArchiveTask call")
	}
	return s.archiveTaskFn(ctx, params)
}

func (s *stubTaskRepository) UpdateTaskStatus(ctx context.Context, params db.UpdateTaskStatusParams) (*db.Task, *apperror.AppError) {
	if s.updateTaskStatusFn == nil {
		panic("unexpected UpdateTaskStatus call")
	}
	return s.updateTaskStatusFn(ctx, params)
}

func (s *stubTaskRepository) UpdateTaskProgress(ctx context.Context, params db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError) {
	if s.updateTaskProgressFn == nil {
		panic("unexpected UpdateTaskProgress call")
	}
	return s.updateTaskProgressFn(ctx, params)
}

func (s *stubTaskRepository) ListTasksByOrganization(ctx context.Context, params db.ListTasksByOrganizationParams) ([]db.Task, *apperror.AppError) {
	if s.listTasksByOrganizationFn == nil {
		panic("unexpected ListTasksByOrganization call")
	}
	return s.listTasksByOrganizationFn(ctx, params)
}

func (s *stubTaskRepository) ListTasksByParent(ctx context.Context, parentTaskID pgtype.UUID) ([]db.Task, *apperror.AppError) {
	if s.listTasksByParentFn == nil {
		panic("unexpected ListTasksByParent call")
	}
	return s.listTasksByParentFn(ctx, parentTaskID)
}

func (s *stubTaskRepository) ListTasksByCreator(ctx context.Context, params db.ListTasksByCreatorParams) ([]db.Task, *apperror.AppError) {
	if s.listTasksByCreatorFn == nil {
		panic("unexpected ListTasksByCreator call")
	}
	return s.listTasksByCreatorFn(ctx, params)
}

func TestNewTaskService(t *testing.T) {
	svc, err := NewTaskService(nil, servicetestutil.NewAllowAuthorizer())
	if err == nil {
		t.Fatal("expected constructor error for nil repository")
	}
	if svc != nil {
		t.Fatal("expected nil service when repository is nil")
	}
	if err.Code != apperror.CodeInternal {
		t.Fatalf("expected internal error, got %s", err.Code)
	}

	svc, err = NewTaskService(&stubTaskRepository{}, servicetestutil.NewAllowAuthorizer())
	if err != nil {
		t.Fatalf("expected constructor success, got error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestCreateTaskValidation(t *testing.T) {
	validUser := &coredto.UserInfo{UserID: uuid.NewString()}

	testCases := []struct {
		name        string
		req         *taskv1.CreateTaskRequest
		userInfo    *coredto.UserInfo
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			userInfo:    validUser,
			wantMessage: "create task request is required",
		},
		{
			name:        "missing user metadata",
			req:         &taskv1.CreateTaskRequest{Title: "Ship feature"},
			userInfo:    nil,
			wantMessage: "user metadata is required",
		},
		{
			name:        "missing title",
			req:         &taskv1.CreateTaskRequest{},
			userInfo:    validUser,
			wantMessage: "title is required",
		},
		{
			name:        "invalid project uuid",
			req:         &taskv1.CreateTaskRequest{Title: "Ship feature", ProjectId: "bad-uuid"},
			userInfo:    validUser,
			wantMessage: "invalid uuid",
		},
		{
			name:        "invalid priority",
			req:         &taskv1.CreateTaskRequest{Title: "Ship feature", Priority: "critical"},
			userInfo:    validUser,
			wantMessage: "invalid task priority",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustTaskService(t, &stubTaskRepository{
				createTaskFn: func(context.Context, db.CreateTaskParams) (*db.Task, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.CreateTask(context.Background(), tc.req, tc.userInfo)
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertTaskAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestCreateTaskSuccess(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	parentTaskID := uuid.New()
	taskID := uuid.New()
	dueAt := time.Date(2026, time.May, 2, 9, 30, 0, 0, time.UTC)

	svc := mustTaskServiceWithAuthorizer(t, &stubTaskRepository{
		createTaskFn: func(_ context.Context, params db.CreateTaskParams) (*db.Task, *apperror.AppError) {
			if params.OrganizationID.Valid {
				t.Fatalf("unexpected organization id: %#v", params.OrganizationID)
			}
			if !params.ProjectID.Valid || params.ProjectID.Bytes != projectID {
				t.Fatalf("unexpected project id: %#v", params.ProjectID)
			}
			if !params.ParentTaskID.Valid || params.ParentTaskID.Bytes != parentTaskID {
				t.Fatalf("unexpected parent task id: %#v", params.ParentTaskID)
			}
			if params.CreatedBy != userID {
				t.Fatalf("unexpected created_by: %s", params.CreatedBy)
			}
			if params.Title != "Ship feature" {
				t.Fatalf("unexpected title: %s", params.Title)
			}
			if params.Description != "Finish the service tests" {
				t.Fatalf("unexpected description: %s", params.Description)
			}
			if params.Priority != "high" {
				t.Fatalf("unexpected priority: %s", params.Priority)
			}
			if !params.DueAt.Valid || !params.DueAt.Time.Equal(dueAt) {
				t.Fatalf("unexpected due_at: %#v", params.DueAt)
			}

			return &db.Task{
				ID:             taskID,
				OrganizationID: params.OrganizationID,
				ProjectID:      params.ProjectID,
				ParentTaskID:   params.ParentTaskID,
				CreatedBy:      params.CreatedBy,
				Title:          params.Title,
				Description:    params.Description,
				Status:         "pending",
				Priority:       params.Priority,
				DueAt:          params.DueAt,
			}, nil
		},
	}, stubTaskAuthorizer{
		requireTaskRoleFn: func(_ context.Context, taskID uuid.UUID, userInfo *dto.UserInfo, minRole authz.Role) (*db.Task, *apperror.AppError) {
			if taskID != parentTaskID {
				t.Fatalf("unexpected parent task id: %s", taskID)
			}
			if userInfo == nil || userInfo.UserID != userID.String() {
				t.Fatalf("unexpected user info: %#v", userInfo)
			}
			if minRole != authz.RoleMember {
				t.Fatalf("unexpected task role requirement: %s", minRole)
			}
			return &db.Task{
				ID:        parentTaskID,
				ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
				CreatedBy: userID,
			}, nil
		},
		requireProjectRoleFn: func(_ context.Context, gotProjectID uuid.UUID, userInfo *dto.UserInfo, minRole authz.Role) (*db.ProjectMembership, *apperror.AppError) {
			if gotProjectID != projectID {
				t.Fatalf("unexpected project id: %s", gotProjectID)
			}
			if userInfo == nil || userInfo.UserID != userID.String() {
				t.Fatalf("unexpected user info: %#v", userInfo)
			}
			if minRole != authz.RoleMember {
				t.Fatalf("unexpected project role requirement: %s", minRole)
			}
			return &db.ProjectMembership{ProjectID: projectID, Role: string(minRole)}, nil
		},
	})

	res, err := svc.CreateTask(context.Background(), &taskv1.CreateTaskRequest{
		ProjectId:    projectID.String(),
		ParentTaskId: parentTaskID.String(),
		Title:        "Ship feature",
		Description:  "Finish the service tests",
		Priority:     "high",
		DueAt:        timestamppb.New(dueAt),
	}, &coredto.UserInfo{UserID: userID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != taskID.String() {
		t.Fatalf("unexpected task id: %s", res.Id)
	}
	if res.ProjectId != projectID.String() {
		t.Fatalf("unexpected project id: %s", res.ProjectId)
	}
	if res.ParentTaskId != parentTaskID.String() {
		t.Fatalf("unexpected parent task id: %s", res.ParentTaskId)
	}
	if res.CreatedBy != userID.String() {
		t.Fatalf("unexpected created_by: %s", res.CreatedBy)
	}
	if res.Priority != "high" {
		t.Fatalf("unexpected priority: %s", res.Priority)
	}
	if res.DueAt == nil || !res.DueAt.AsTime().Equal(dueAt) {
		t.Fatalf("unexpected due_at: %#v", res.DueAt)
	}
}

func TestCreateTaskRejectsMismatchedParentProjectScope(t *testing.T) {
	parentTaskID := uuid.New()
	parentProjectID := uuid.New()
	requestProjectID := uuid.New()

	svc := mustTaskServiceWithAuthorizer(t, &stubTaskRepository{
		createTaskFn: func(context.Context, db.CreateTaskParams) (*db.Task, *apperror.AppError) {
			t.Fatal("repository should not be called for mismatched parent scope")
			return nil, nil
		},
	}, stubTaskAuthorizer{
		requireTaskRoleFn: func(_ context.Context, taskID uuid.UUID, _ *dto.UserInfo, minRole authz.Role) (*db.Task, *apperror.AppError) {
			if taskID != parentTaskID {
				t.Fatalf("unexpected parent task id: %s", taskID)
			}
			if minRole != authz.RoleMember {
				t.Fatalf("unexpected role requirement: %s", minRole)
			}
			return &db.Task{
				ID:        parentTaskID,
				ProjectID: pgtype.UUID{Bytes: parentProjectID, Valid: true},
				CreatedBy: uuid.New(),
			}, nil
		},
		requireProjectRoleFn: func(context.Context, uuid.UUID, *dto.UserInfo, authz.Role) (*db.ProjectMembership, *apperror.AppError) {
			t.Fatal("project authorization should not run after a parent scope mismatch")
			return nil, nil
		},
	})

	res, err := svc.CreateTask(context.Background(), &taskv1.CreateTaskRequest{
		ProjectId:    requestProjectID.String(),
		ParentTaskId: parentTaskID.String(),
		Title:        "Ship feature",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAppError(t, err, apperror.CodeValidation, "child task scope must match parent task scope")
}

func TestGetTaskRepoError(t *testing.T) {
	taskID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeNotFound, Message: "task not found"}

	svc := mustTaskService(t, &stubTaskRepository{
		getTaskFn: func(_ context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
			if id != taskID {
				t.Fatalf("unexpected task id: %s", id)
			}
			return nil, repoErr
		},
	})

	res, err := svc.GetTask(context.Background(), &taskv1.GetTaskRequest{Id: taskID.String()}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestListTasksByProjectFiltersByStatus(t *testing.T) {
	projectID := uuid.New()
	pendingID := uuid.New()
	completedID := uuid.New()
	creatorID := uuid.New()

	svc := mustTaskService(t, &stubTaskRepository{
		listTasksByProjectFn: func(_ context.Context, got pgtype.UUID) ([]db.Task, *apperror.AppError) {
			if !got.Valid || got.Bytes != projectID {
				t.Fatalf("unexpected project id: %#v", got)
			}

			return []db.Task{
				{ID: pendingID, CreatedBy: creatorID, Title: "Pending", Status: "pending", Priority: "medium"},
				{ID: completedID, CreatedBy: creatorID, Title: "Done", Status: "completed", Priority: "high"},
			}, nil
		},
	})

	res, err := svc.ListTasksByProject(context.Background(), &taskv1.ListTasksByProjectRequest{
		ProjectId: projectID.String(),
		Status:    "completed",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(res.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(res.Tasks))
	}
	if res.Tasks[0].Id != completedID.String() {
		t.Fatalf("unexpected task id: %s", res.Tasks[0].Id)
	}
}

func TestListTasksByProjectValidationAndRepoError(t *testing.T) {
	svc := mustTaskService(t, &stubTaskRepository{
		listTasksByProjectFn: func(context.Context, pgtype.UUID) ([]db.Task, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid request")
			return nil, nil
		},
	})

	res, err := svc.ListTasksByProject(context.Background(), &taskv1.ListTasksByProjectRequest{
		ProjectId: uuid.NewString(),
		Status:    "done",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAppError(t, err, apperror.CodeValidation, "invalid task status")

	projectID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to list tasks by project"}
	svc = mustTaskService(t, &stubTaskRepository{
		listTasksByProjectFn: func(_ context.Context, got pgtype.UUID) ([]db.Task, *apperror.AppError) {
			if !got.Valid || got.Bytes != projectID {
				t.Fatalf("unexpected project id: %#v", got)
			}
			return nil, repoErr
		},
	})

	res, err = svc.ListTasksByProject(context.Background(), &taskv1.ListTasksByProjectRequest{
		ProjectId: projectID.String(),
		Status:    "pending",
	}, &dto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestUpdateTaskStatusCompletedSetsStartedAndCompletedAt(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()

	getTaskCalled := false
	updateCalled := false

	svc := mustTaskService(t, &stubTaskRepository{
		getTaskFn: func(_ context.Context, got uuid.UUID) (*db.Task, *apperror.AppError) {
			getTaskCalled = true
			if got != taskID {
				t.Fatalf("unexpected task id: %s", got)
			}
			return &db.Task{
				ID:        taskID,
				CreatedBy: uuid.New(),
				Title:     "Ship feature",
				Status:    "pending",
				Priority:  "medium",
			}, nil
		},
		updateTaskStatusFn: func(_ context.Context, params db.UpdateTaskStatusParams) (*db.Task, *apperror.AppError) {
			updateCalled = true
			if params.ID != taskID {
				t.Fatalf("unexpected task id: %s", params.ID)
			}
			if !params.UpdatedBy.Valid || params.UpdatedBy.Bytes != userID {
				t.Fatalf("unexpected updated_by: %#v", params.UpdatedBy)
			}
			if params.Status != "completed" {
				t.Fatalf("unexpected status: %s", params.Status)
			}
			if !params.StartedAt.Valid {
				t.Fatal("expected started_at to be set")
			}
			if !params.CompletedAt.Valid {
				t.Fatal("expected completed_at to be set")
			}
			if params.CompletedAt.Time.Before(params.StartedAt.Time) {
				t.Fatalf("expected completed_at >= started_at, got %v < %v", params.CompletedAt.Time, params.StartedAt.Time)
			}

			return &db.Task{
				ID:          taskID,
				CreatedBy:   uuid.New(),
				UpdatedBy:   params.UpdatedBy,
				Title:       "Ship feature",
				Status:      params.Status,
				Priority:    "medium",
				StartedAt:   params.StartedAt,
				CompletedAt: params.CompletedAt,
			}, nil
		},
	})

	res, err := svc.UpdateTaskStatus(context.Background(), &taskv1.UpdateTaskStatusRequest{
		Id:     taskID.String(),
		Status: "completed",
	}, &coredto.UserInfo{UserID: userID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !getTaskCalled {
		t.Fatal("expected GetTask to be called")
	}
	if !updateCalled {
		t.Fatal("expected UpdateTaskStatus to be called")
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Status != "completed" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
	if res.StartedAt == nil {
		t.Fatal("expected started_at in response")
	}
	if res.CompletedAt == nil {
		t.Fatal("expected completed_at in response")
	}
}

func TestUpdateTaskStatusInvalidStatus(t *testing.T) {
	svc := mustTaskService(t, &stubTaskRepository{
		getTaskFn: func(context.Context, uuid.UUID) (*db.Task, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid status")
			return nil, nil
		},
	})

	res, err := svc.UpdateTaskStatus(context.Background(), &taskv1.UpdateTaskStatusRequest{
		Id:     uuid.NewString(),
		Status: "done",
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAppError(t, err, apperror.CodeValidation, "invalid task status")
}

func TestUpdateTaskProgressRejectsOutOfRange(t *testing.T) {
	testCases := []int32{-1, 101}

	for _, progress := range testCases {
		t.Run(fmt.Sprintf("progress_%d", progress), func(t *testing.T) {
			repoCalled := false
			svc := mustTaskService(t, &stubTaskRepository{
				getTaskFn: func(context.Context, uuid.UUID) (*db.Task, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
				updateTaskProgressFn: func(context.Context, db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.UpdateTaskProgress(context.Background(), &taskv1.UpdateTaskProgressRequest{
				Id:              uuid.NewString(),
				ProgressPercent: progress,
			}, &coredto.UserInfo{UserID: uuid.NewString()})
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertTaskAppError(t, err, apperror.CodeValidation, "progress_percent must be between 0 and 100")
			if repoCalled {
				t.Fatal("expected repository not to be called for invalid progress")
			}
		})
	}
}

func TestUpdateTaskProgressRepoError(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to update task progress"}

	svc := mustTaskService(t, &stubTaskRepository{
		getTaskFn: func(_ context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
			if id != taskID {
				t.Fatalf("unexpected task id: %s", id)
			}
			return &db.Task{ID: taskID, CreatedBy: uuid.New(), Title: "Ship feature", Priority: "medium"}, nil
		},
		updateTaskProgressFn: func(_ context.Context, params db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError) {
			if params.ID != taskID {
				t.Fatalf("unexpected task id: %s", params.ID)
			}
			if !params.UpdatedBy.Valid || params.UpdatedBy.Bytes != userID {
				t.Fatalf("unexpected updated_by: %#v", params.UpdatedBy)
			}
			if params.ProgressPercent != 55 {
				t.Fatalf("unexpected progress: %d", params.ProgressPercent)
			}
			return nil, repoErr
		},
	})

	res, err := svc.UpdateTaskProgress(context.Background(), &taskv1.UpdateTaskProgressRequest{
		Id:              taskID.String(),
		ProgressPercent: 55,
	}, &coredto.UserInfo{UserID: userID.String()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestListTasksByOrganizationAndCreatorValidation(t *testing.T) {
	svc := mustTaskService(t, &stubTaskRepository{
		listTasksByOrganizationFn: func(context.Context, db.ListTasksByOrganizationParams) ([]db.Task, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid organization id")
			return nil, nil
		},
		listTasksByCreatorFn: func(context.Context, db.ListTasksByCreatorParams) ([]db.Task, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid creator id")
			return nil, nil
		},
	})

	orgRes, orgErr := svc.ListTasksByOrganization(context.Background(), &taskv1.ListTasksByOrganizationRequest{
		OrganizationId: "bad-uuid",
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if orgRes != nil {
		t.Fatalf("expected nil response, got %#v", orgRes)
	}
	assertTaskAppError(t, orgErr, apperror.CodeValidation, "invalid uuid")

	creatorRes, creatorErr := svc.ListTasksByCreator(context.Background(), &taskv1.ListTasksByCreatorRequest{
		CreatedBy: "bad-uuid",
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if creatorRes != nil {
		t.Fatalf("expected nil response, got %#v", creatorRes)
	}
	assertTaskAppError(t, creatorErr, apperror.CodeValidation, "invalid uuid")
}

func TestListTasksByParentFiltersMismatchedChildren(t *testing.T) {
	parentTaskID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	keptTaskID := uuid.New()
	wrongProjectTaskID := uuid.New()
	personalTaskID := uuid.New()

	svc := mustTaskServiceWithAuthorizer(t, &stubTaskRepository{
		listTasksByParentFn: func(_ context.Context, parent pgtype.UUID) ([]db.Task, *apperror.AppError) {
			if !parent.Valid || parent.Bytes != parentTaskID {
				t.Fatalf("unexpected parent task id: %#v", parent)
			}
			return []db.Task{
				{
					ID:        keptTaskID,
					ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
					CreatedBy: userID,
					Title:     "Kept",
					Priority:  "medium",
				},
				{
					ID:        wrongProjectTaskID,
					ProjectID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
					CreatedBy: userID,
					Title:     "Wrong project",
					Priority:  "medium",
				},
				{
					ID:        personalTaskID,
					CreatedBy: userID,
					Title:     "Wrong scope",
					Priority:  "medium",
				},
			}, nil
		},
	}, stubTaskAuthorizer{
		requireTaskRoleFn: func(_ context.Context, taskID uuid.UUID, gotUserInfo *dto.UserInfo, minRole authz.Role) (*db.Task, *apperror.AppError) {
			if taskID != parentTaskID {
				t.Fatalf("unexpected parent task id: %s", taskID)
			}
			if gotUserInfo == nil || gotUserInfo.UserID != userID.String() {
				t.Fatalf("unexpected user info: %#v", gotUserInfo)
			}
			if minRole != authz.RoleMember {
				t.Fatalf("unexpected role requirement: %s", minRole)
			}
			return &db.Task{
				ID:        parentTaskID,
				ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
				CreatedBy: userID,
			}, nil
		},
	})

	res, err := svc.ListTasksByParent(context.Background(), &taskv1.ListTasksByParentRequest{
		ParentTaskId: parentTaskID.String(),
	}, &dto.UserInfo{UserID: userID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(res.Tasks) != 1 {
		t.Fatalf("expected 1 task after filtering, got %d", len(res.Tasks))
	}
	if res.Tasks[0].Id != keptTaskID.String() {
		t.Fatalf("unexpected task kept after filtering: %s", res.Tasks[0].Id)
	}
}

func mustTaskService(t *testing.T, repo *stubTaskRepository) TaskService {
	t.Helper()

	return mustTaskServiceWithAuthorizer(t, repo, servicetestutil.NewAllowAuthorizer())
}

func mustTaskServiceWithAuthorizer(t *testing.T, repo *stubTaskRepository, authorizer authz.Authorizer) TaskService {
	t.Helper()

	svc, err := NewTaskService(repo, authorizer)
	if err != nil {
		t.Fatalf("failed to construct task service: %v", err)
	}

	return svc
}

func assertTaskAppError(t *testing.T, err *apperror.AppError, code apperror.ErrorCode, message string) {
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
