package taskassigment

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type stubTaskAssignmentRepository struct {
	assignTaskFn                  func(context.Context, db.AssignTaskParams) (*db.TaskAssignment, *apperror.AppError)
	getActiveTaskAssignmentFn     func(context.Context, db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError)
	unassignTaskFn                func(context.Context, db.UnassignTaskParams) (*db.TaskAssignment, *apperror.AppError)
	listTaskAssignmentsFn         func(context.Context, uuid.UUID) ([]db.TaskAssignment, *apperror.AppError)
	listActiveAssignmentsByUserFn func(context.Context, db.ListActiveAssignmentsByAssigneeParams) ([]db.TaskAssignment, *apperror.AppError)
}

func (s *stubTaskAssignmentRepository) AssignTask(ctx context.Context, params db.AssignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
	if s.assignTaskFn == nil {
		panic("unexpected AssignTask call")
	}
	return s.assignTaskFn(ctx, params)
}

func (s *stubTaskAssignmentRepository) GetActiveTaskAssignment(ctx context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
	if s.getActiveTaskAssignmentFn == nil {
		panic("unexpected GetActiveTaskAssignment call")
	}
	return s.getActiveTaskAssignmentFn(ctx, params)
}

func (s *stubTaskAssignmentRepository) UnassignTask(ctx context.Context, params db.UnassignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
	if s.unassignTaskFn == nil {
		panic("unexpected UnassignTask call")
	}
	return s.unassignTaskFn(ctx, params)
}

func (s *stubTaskAssignmentRepository) ListTaskAssignments(ctx context.Context, taskID uuid.UUID) ([]db.TaskAssignment, *apperror.AppError) {
	if s.listTaskAssignmentsFn == nil {
		panic("unexpected ListTaskAssignments call")
	}
	return s.listTaskAssignmentsFn(ctx, taskID)
}

func (s *stubTaskAssignmentRepository) ListActiveAssignmentsByAssignee(ctx context.Context, params db.ListActiveAssignmentsByAssigneeParams) ([]db.TaskAssignment, *apperror.AppError) {
	if s.listActiveAssignmentsByUserFn == nil {
		panic("unexpected ListActiveAssignmentsByAssignee call")
	}
	return s.listActiveAssignmentsByUserFn(ctx, params)
}

func TestNewTaskAssignmentService(t *testing.T) {
	svc, err := NewTaskAssignmentService(nil)
	if err == nil {
		t.Fatal("expected constructor error for nil repository")
	}
	if svc != nil {
		t.Fatal("expected nil service when repository is nil")
	}
	if err.Code != apperror.CodeInternal {
		t.Fatalf("expected internal error, got %s", err.Code)
	}

	svc, err = NewTaskAssignmentService(&stubTaskAssignmentRepository{})
	if err != nil {
		t.Fatalf("expected constructor success, got error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestAssignTaskValidation(t *testing.T) {
	validUser := &coredto.UserInfo{UserID: uuid.NewString()}

	testCases := []struct {
		name        string
		req         *taskv1.AssignTaskRequest
		userInfo    *coredto.UserInfo
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			userInfo:    validUser,
			wantMessage: "assign task request is required",
		},
		{
			name:        "missing user metadata",
			req:         &taskv1.AssignTaskRequest{TaskId: uuid.NewString(), AssigneeType: "user", AssigneeId: uuid.NewString()},
			userInfo:    nil,
			wantMessage: "user metadata is required",
		},
		{
			name:        "missing task id",
			req:         &taskv1.AssignTaskRequest{AssigneeType: "user", AssigneeId: uuid.NewString()},
			userInfo:    validUser,
			wantMessage: "task id is required",
		},
		{
			name:        "invalid assignee type",
			req:         &taskv1.AssignTaskRequest{TaskId: uuid.NewString(), AssigneeType: "group", AssigneeId: uuid.NewString()},
			userInfo:    validUser,
			wantMessage: "invalid assignee type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
				getActiveTaskAssignmentFn: func(context.Context, db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
				assignTaskFn: func(context.Context, db.AssignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.AssignTask(context.Background(), tc.req, tc.userInfo)
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertTaskAssignmentAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestAssignTaskDuplicateConflict(t *testing.T) {
	taskID := uuid.New()
	assigneeID := uuid.New()

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		getActiveTaskAssignmentFn: func(_ context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
			if params.TaskID != taskID || params.AssigneeID != assigneeID || params.AssigneeType != "user" {
				t.Fatalf("unexpected lookup params: %#v", params)
			}
			return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "user", AssigneeID: assigneeID}, nil
		},
	})

	res, err := svc.AssignTask(context.Background(), &taskv1.AssignTaskRequest{
		TaskId:       taskID.String(),
		AssigneeType: "user",
		AssigneeId:   assigneeID.String(),
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAssignmentAppError(t, err, apperror.CodeConflict, "task assignment already exists")
}

func TestAssignTaskSuccess(t *testing.T) {
	taskID := uuid.New()
	assigneeID := uuid.New()
	assignedBy := uuid.New()
	assignmentID := uuid.New()
	assignedAt := pgtype.Timestamptz{Time: time.Date(2026, time.May, 4, 12, 0, 0, 0, time.UTC), Valid: true}

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		getActiveTaskAssignmentFn: func(_ context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
			if params.TaskID != taskID || params.AssigneeID != assigneeID || params.AssigneeType != "user" {
				t.Fatalf("unexpected lookup params: %#v", params)
			}
			return nil, &apperror.AppError{Code: apperror.CodeNotFound, Message: "task assignment not found"}
		},
		assignTaskFn: func(_ context.Context, params db.AssignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			if params.TaskID != taskID {
				t.Fatalf("unexpected task id: %s", params.TaskID)
			}
			if params.AssigneeType != "user" {
				t.Fatalf("unexpected assignee type: %s", params.AssigneeType)
			}
			if params.AssigneeID != assigneeID {
				t.Fatalf("unexpected assignee id: %s", params.AssigneeID)
			}
			if params.AssignedBy != assignedBy {
				t.Fatalf("unexpected assigned_by: %s", params.AssignedBy)
			}
			return &db.TaskAssignment{
				ID:           assignmentID,
				TaskID:       taskID,
				AssigneeType: params.AssigneeType,
				AssigneeID:   params.AssigneeID,
				AssignedBy:   params.AssignedBy,
				AssignedAt:   assignedAt,
			}, nil
		},
	})

	res, err := svc.AssignTask(context.Background(), &taskv1.AssignTaskRequest{
		TaskId:       taskID.String(),
		AssigneeType: "user",
		AssigneeId:   assigneeID.String(),
	}, &coredto.UserInfo{UserID: assignedBy.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != assignmentID.String() {
		t.Fatalf("unexpected id: %s", res.Id)
	}
	if res.TaskId != taskID.String() {
		t.Fatalf("unexpected task id: %s", res.TaskId)
	}
	if res.AssigneeId != assigneeID.String() {
		t.Fatalf("unexpected assignee id: %s", res.AssigneeId)
	}
}

func TestUnassignTaskSuccess(t *testing.T) {
	taskID := uuid.New()
	assigneeID := uuid.New()

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		unassignTaskFn: func(_ context.Context, params db.UnassignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			if params.TaskID != taskID {
				t.Fatalf("unexpected task id: %s", params.TaskID)
			}
			if params.AssigneeType != "team" {
				t.Fatalf("unexpected assignee type: %s", params.AssigneeType)
			}
			if params.AssigneeID != assigneeID {
				t.Fatalf("unexpected assignee id: %s", params.AssigneeID)
			}
			return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "team", AssigneeID: assigneeID}, nil
		},
	})

	res, err := svc.UnassignTask(context.Background(), &taskv1.UnassignTaskRequest{
		TaskId:       taskID.String(),
		AssigneeType: "team",
		AssigneeId:   assigneeID.String(),
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("expected success response, got %#v", res)
	}
}

func TestUnassignTaskValidation(t *testing.T) {
	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		unassignTaskFn: func(context.Context, db.UnassignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid request")
			return nil, nil
		},
	})

	res, err := svc.UnassignTask(context.Background(), &taskv1.UnassignTaskRequest{
		TaskId:       uuid.NewString(),
		AssigneeType: "group",
		AssigneeId:   uuid.NewString(),
	})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAssignmentAppError(t, err, apperror.CodeValidation, "invalid assignee type")
}

func TestReassignTaskSameTargetValidation(t *testing.T) {
	taskID := uuid.New()
	assigneeID := uuid.New()

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{})
	res, err := svc.ReassignTask(context.Background(), &taskv1.ReassignTaskRequest{
		TaskId:           taskID.String(),
		FromAssigneeType: "user",
		FromAssigneeId:   assigneeID.String(),
		ToAssigneeType:   "user",
		ToAssigneeId:     assigneeID.String(),
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAssignmentAppError(t, err, apperror.CodeValidation, "reassignment target must be different from the current assignee")
}

func TestReassignTaskTargetConflict(t *testing.T) {
	taskID := uuid.New()
	fromAssigneeID := uuid.New()
	toAssigneeID := uuid.New()
	lookupCalls := 0

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		getActiveTaskAssignmentFn: func(_ context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
			lookupCalls++
			switch lookupCalls {
			case 1:
				return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "user", AssigneeID: fromAssigneeID}, nil
			case 2:
				return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "team", AssigneeID: toAssigneeID}, nil
			default:
				t.Fatalf("unexpected extra lookup: %#v", params)
				return nil, nil
			}
		},
	})

	res, err := svc.ReassignTask(context.Background(), &taskv1.ReassignTaskRequest{
		TaskId:           taskID.String(),
		FromAssigneeType: "user",
		FromAssigneeId:   fromAssigneeID.String(),
		ToAssigneeType:   "team",
		ToAssigneeId:     toAssigneeID.String(),
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAssignmentAppError(t, err, apperror.CodeConflict, "task assignment already exists")
}

func TestReassignTaskSuccess(t *testing.T) {
	taskID := uuid.New()
	fromAssigneeID := uuid.New()
	toAssigneeID := uuid.New()
	assignedBy := uuid.New()
	assignmentID := uuid.New()

	lookupCalls := 0
	unassignCalls := 0

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		getActiveTaskAssignmentFn: func(_ context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
			lookupCalls++
			switch lookupCalls {
			case 1:
				if params.TaskID != taskID || params.AssigneeID != fromAssigneeID || params.AssigneeType != "user" {
					t.Fatalf("unexpected source lookup params: %#v", params)
				}
				return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "user", AssigneeID: fromAssigneeID}, nil
			case 2:
				if params.TaskID != taskID || params.AssigneeID != toAssigneeID || params.AssigneeType != "team" {
					t.Fatalf("unexpected target lookup params: %#v", params)
				}
				return nil, &apperror.AppError{Code: apperror.CodeNotFound, Message: "task assignment not found"}
			default:
				t.Fatalf("unexpected extra lookup: %#v", params)
				return nil, nil
			}
		},
		assignTaskFn: func(_ context.Context, params db.AssignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			if params.TaskID != taskID || params.AssigneeType != "team" || params.AssigneeID != toAssigneeID || params.AssignedBy != assignedBy {
				t.Fatalf("unexpected assign params: %#v", params)
			}
			return &db.TaskAssignment{
				ID:           assignmentID,
				TaskID:       taskID,
				AssigneeType: "team",
				AssigneeID:   toAssigneeID,
				AssignedBy:   assignedBy,
			}, nil
		},
		unassignTaskFn: func(_ context.Context, params db.UnassignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			unassignCalls++
			if unassignCalls != 1 {
				t.Fatalf("unexpected unassign call count: %d", unassignCalls)
			}
			if params.TaskID != taskID || params.AssigneeType != "user" || params.AssigneeID != fromAssigneeID {
				t.Fatalf("unexpected unassign params: %#v", params)
			}
			return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "user", AssigneeID: fromAssigneeID}, nil
		},
	})

	res, err := svc.ReassignTask(context.Background(), &taskv1.ReassignTaskRequest{
		TaskId:           taskID.String(),
		FromAssigneeType: "user",
		FromAssigneeId:   fromAssigneeID.String(),
		ToAssigneeType:   "team",
		ToAssigneeId:     toAssigneeID.String(),
	}, &coredto.UserInfo{UserID: assignedBy.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != assignmentID.String() {
		t.Fatalf("unexpected id: %s", res.Id)
	}
	if res.AssigneeType != "team" {
		t.Fatalf("unexpected assignee type: %s", res.AssigneeType)
	}
	if res.AssigneeId != toAssigneeID.String() {
		t.Fatalf("unexpected assignee id: %s", res.AssigneeId)
	}
}

func TestReassignTaskRollbackFailurePath(t *testing.T) {
	taskID := uuid.New()
	fromAssigneeID := uuid.New()
	toAssigneeID := uuid.New()
	assignedBy := uuid.New()
	lookupCalls := 0
	unassignCalls := 0
	sourceErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to unassign task"}

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		getActiveTaskAssignmentFn: func(_ context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError) {
			lookupCalls++
			switch lookupCalls {
			case 1:
				return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "user", AssigneeID: fromAssigneeID}, nil
			case 2:
				return nil, &apperror.AppError{Code: apperror.CodeNotFound, Message: "task assignment not found"}
			default:
				t.Fatalf("unexpected extra lookup: %#v", params)
				return nil, nil
			}
		},
		assignTaskFn: func(_ context.Context, params db.AssignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			if params.TaskID != taskID || params.AssigneeID != toAssigneeID || params.AssigneeType != "team" || params.AssignedBy != assignedBy {
				t.Fatalf("unexpected assign params: %#v", params)
			}
			return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "team", AssigneeID: toAssigneeID, AssignedBy: assignedBy}, nil
		},
		unassignTaskFn: func(_ context.Context, params db.UnassignTaskParams) (*db.TaskAssignment, *apperror.AppError) {
			unassignCalls++
			switch unassignCalls {
			case 1:
				if params.TaskID != taskID || params.AssigneeID != fromAssigneeID || params.AssigneeType != "user" {
					t.Fatalf("unexpected source unassign params: %#v", params)
				}
				return nil, sourceErr
			case 2:
				if params.TaskID != taskID || params.AssigneeID != toAssigneeID || params.AssigneeType != "team" {
					t.Fatalf("unexpected rollback unassign params: %#v", params)
				}
				return &db.TaskAssignment{ID: uuid.New(), TaskID: taskID, AssigneeType: "team", AssigneeID: toAssigneeID}, nil
			default:
				t.Fatalf("unexpected unassign call count: %d", unassignCalls)
				return nil, nil
			}
		},
	})

	res, err := svc.ReassignTask(context.Background(), &taskv1.ReassignTaskRequest{
		TaskId:           taskID.String(),
		FromAssigneeType: "user",
		FromAssigneeId:   fromAssigneeID.String(),
		ToAssigneeType:   "team",
		ToAssigneeId:     toAssigneeID.String(),
	}, &coredto.UserInfo{UserID: assignedBy.String()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != sourceErr {
		t.Fatalf("expected original unassign error, got %#v", err)
	}
	if unassignCalls != 2 {
		t.Fatalf("expected rollback unassign to be attempted, got %d unassign calls", unassignCalls)
	}
}

func TestListTaskAssignmentsSuccess(t *testing.T) {
	taskID := uuid.New()
	assignmentID := uuid.New()
	assigneeID := uuid.New()

	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		listTaskAssignmentsFn: func(_ context.Context, got uuid.UUID) ([]db.TaskAssignment, *apperror.AppError) {
			if got != taskID {
				t.Fatalf("unexpected task id: %s", got)
			}
			return []db.TaskAssignment{
				{
					ID:           assignmentID,
					TaskID:       taskID,
					AssigneeType: "user",
					AssigneeID:   assigneeID,
					AssignedBy:   uuid.New(),
				},
			}, nil
		},
	})

	res, err := svc.ListTaskAssignments(context.Background(), &taskv1.ListTaskAssignmentsRequest{
		TaskId: taskID.String(),
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(res.Assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(res.Assignments))
	}
	if res.Assignments[0].Id != assignmentID.String() {
		t.Fatalf("unexpected assignment id: %s", res.Assignments[0].Id)
	}
}

func TestListTaskAssignmentsValidation(t *testing.T) {
	svc := mustTaskAssignmentService(t, &stubTaskAssignmentRepository{
		listTaskAssignmentsFn: func(context.Context, uuid.UUID) ([]db.TaskAssignment, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid task id")
			return nil, nil
		},
	})

	res, err := svc.ListTaskAssignments(context.Background(), &taskv1.ListTaskAssignmentsRequest{
		TaskId: "bad-uuid",
	})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskAssignmentAppError(t, err, apperror.CodeValidation, "invalid uuid")
}

func mustTaskAssignmentService(t *testing.T, repo *stubTaskAssignmentRepository) TaskAssignmentService {
	t.Helper()

	svc, err := NewTaskAssignmentService(repo)
	if err != nil {
		t.Fatalf("failed to construct task assignment service: %v", err)
	}

	return svc
}

func assertTaskAssignmentAppError(t *testing.T, err *apperror.AppError, code apperror.ErrorCode, message string) {
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
