package taskcomment

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
	servicetestutil "github.com/rijum8906/relay/services/task-service/internal/services/testutil"
)

type stubTaskCommentRepository struct {
	createTaskCommentFn func(context.Context, db.CreateTaskCommentParams) (*db.TaskComment, *apperror.AppError)
	getTaskCommentFn    func(context.Context, uuid.UUID) (*db.TaskComment, *apperror.AppError)
	updateTaskCommentFn func(context.Context, db.UpdateTaskCommentParams) (*db.TaskComment, *apperror.AppError)
	deleteTaskCommentFn func(context.Context, db.DeleteTaskCommentParams) (*db.TaskComment, *apperror.AppError)
	listTaskCommentsFn  func(context.Context, uuid.UUID) ([]db.TaskComment, *apperror.AppError)
}

func (s *stubTaskCommentRepository) CreateTaskComment(ctx context.Context, params db.CreateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
	if s.createTaskCommentFn == nil {
		panic("unexpected CreateTaskComment call")
	}
	return s.createTaskCommentFn(ctx, params)
}

func (s *stubTaskCommentRepository) GetTaskComment(ctx context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
	if s.getTaskCommentFn == nil {
		panic("unexpected GetTaskComment call")
	}
	return s.getTaskCommentFn(ctx, id)
}

func (s *stubTaskCommentRepository) UpdateTaskComment(ctx context.Context, params db.UpdateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
	if s.updateTaskCommentFn == nil {
		panic("unexpected UpdateTaskComment call")
	}
	return s.updateTaskCommentFn(ctx, params)
}

func (s *stubTaskCommentRepository) DeleteTaskComment(ctx context.Context, params db.DeleteTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
	if s.deleteTaskCommentFn == nil {
		panic("unexpected DeleteTaskComment call")
	}
	return s.deleteTaskCommentFn(ctx, params)
}

func (s *stubTaskCommentRepository) ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]db.TaskComment, *apperror.AppError) {
	if s.listTaskCommentsFn == nil {
		panic("unexpected ListTaskComments call")
	}
	return s.listTaskCommentsFn(ctx, taskID)
}

func TestNewTaskCommentService(t *testing.T) {
	svc, err := NewTaskCommentService(nil, servicetestutil.NewAllowAuthorizer())
	if err == nil {
		t.Fatal("expected constructor error for nil repository")
	}
	if svc != nil {
		t.Fatal("expected nil service when repository is nil")
	}
	if err.Code != apperror.CodeInternal {
		t.Fatalf("expected internal error, got %s", err.Code)
	}

	svc, err = NewTaskCommentService(&stubTaskCommentRepository{}, servicetestutil.NewAllowAuthorizer())
	if err != nil {
		t.Fatalf("expected constructor success, got error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestCreateTaskCommentValidation(t *testing.T) {
	validUser := &coredto.UserInfo{UserID: uuid.NewString()}

	testCases := []struct {
		name        string
		req         *taskv1.CreateTaskCommentRequest
		userInfo    *coredto.UserInfo
		wantMessage string
	}{
		{
			name:        "nil request",
			req:         nil,
			userInfo:    validUser,
			wantMessage: "create task comment request is required",
		},
		{
			name:        "missing user metadata",
			req:         &taskv1.CreateTaskCommentRequest{TaskId: uuid.NewString(), Body: "Looks good"},
			userInfo:    nil,
			wantMessage: "user metadata is required",
		},
		{
			name:        "missing body",
			req:         &taskv1.CreateTaskCommentRequest{TaskId: uuid.NewString()},
			userInfo:    validUser,
			wantMessage: "body is required",
		},
		{
			name:        "invalid task id",
			req:         &taskv1.CreateTaskCommentRequest{TaskId: "bad-uuid", Body: "Looks good"},
			userInfo:    validUser,
			wantMessage: "invalid uuid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			svc := mustTaskCommentService(t, &stubTaskCommentRepository{
				createTaskCommentFn: func(context.Context, db.CreateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
					repoCalled = true
					return nil, nil
				},
			})

			res, err := svc.CreateTaskComment(context.Background(), tc.req, tc.userInfo)
			if res != nil {
				t.Fatalf("expected nil response, got %#v", res)
			}
			assertTaskCommentAppError(t, err, apperror.CodeValidation, tc.wantMessage)
			if repoCalled {
				t.Fatal("expected repository not to be called for validation failure")
			}
		})
	}
}

func TestCreateTaskCommentSuccess(t *testing.T) {
	taskID := uuid.New()
	authorID := uuid.New()
	commentID := uuid.New()
	createdAt := pgtype.Timestamptz{Time: time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC), Valid: true}

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		createTaskCommentFn: func(_ context.Context, params db.CreateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
			if params.TaskID != taskID {
				t.Fatalf("unexpected task id: %s", params.TaskID)
			}
			if params.AuthorID != authorID {
				t.Fatalf("unexpected author id: %s", params.AuthorID)
			}
			if params.Body != "Looks good" {
				t.Fatalf("unexpected body: %s", params.Body)
			}
			return &db.TaskComment{
				ID:        commentID,
				TaskID:    taskID,
				AuthorID:  authorID,
				Body:      params.Body,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			}, nil
		},
	})

	res, err := svc.CreateTaskComment(context.Background(), &taskv1.CreateTaskCommentRequest{
		TaskId: taskID.String(),
		Body:   "Looks good",
	}, &coredto.UserInfo{UserID: authorID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Id != commentID.String() {
		t.Fatalf("unexpected id: %s", res.Id)
	}
	if res.TaskId != taskID.String() {
		t.Fatalf("unexpected task id: %s", res.TaskId)
	}
	if res.AuthorId != authorID.String() {
		t.Fatalf("unexpected author id: %s", res.AuthorId)
	}
}

func TestUpdateTaskCommentForbiddenForNonAuthor(t *testing.T) {
	commentID := uuid.New()
	authorID := uuid.New()
	otherUserID := uuid.New()

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		getTaskCommentFn: func(_ context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
			if id != commentID {
				t.Fatalf("unexpected comment id: %s", id)
			}
			return &db.TaskComment{
				ID:       commentID,
				TaskID:   uuid.New(),
				AuthorID: authorID,
				Body:     "Original",
			}, nil
		},
	})

	res, err := svc.UpdateTaskComment(context.Background(), &taskv1.UpdateTaskCommentRequest{
		Id:   commentID.String(),
		Body: "Edited",
	}, &coredto.UserInfo{UserID: otherUserID.String()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskCommentAppError(t, err, apperror.CodeForbidden, "only the comment author can update this comment")
}

func TestUpdateTaskCommentSuccess(t *testing.T) {
	commentID := uuid.New()
	taskID := uuid.New()
	authorID := uuid.New()
	editedAt := pgtype.Timestamptz{Time: time.Date(2026, time.May, 4, 11, 0, 0, 0, time.UTC), Valid: true}

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		getTaskCommentFn: func(_ context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
			if id != commentID {
				t.Fatalf("unexpected comment id: %s", id)
			}
			return &db.TaskComment{
				ID:       commentID,
				TaskID:   taskID,
				AuthorID: authorID,
				Body:     "Original",
			}, nil
		},
		updateTaskCommentFn: func(_ context.Context, params db.UpdateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
			if params.ID != commentID {
				t.Fatalf("unexpected comment id: %s", params.ID)
			}
			if params.Body != "Edited" {
				t.Fatalf("unexpected body: %s", params.Body)
			}
			return &db.TaskComment{
				ID:        commentID,
				TaskID:    taskID,
				AuthorID:  authorID,
				Body:      params.Body,
				EditedAt:  editedAt,
				EditCount: 1,
			}, nil
		},
	})

	res, err := svc.UpdateTaskComment(context.Background(), &taskv1.UpdateTaskCommentRequest{
		Id:   commentID.String(),
		Body: "Edited",
	}, &coredto.UserInfo{UserID: authorID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.Body != "Edited" {
		t.Fatalf("unexpected body: %s", res.Body)
	}
	if res.EditCount != 1 {
		t.Fatalf("unexpected edit count: %d", res.EditCount)
	}
}

func TestUpdateTaskCommentGetRepoError(t *testing.T) {
	commentID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeNotFound, Message: "task comment not found"}

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		getTaskCommentFn: func(_ context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
			if id != commentID {
				t.Fatalf("unexpected comment id: %s", id)
			}
			return nil, repoErr
		},
	})

	res, err := svc.UpdateTaskComment(context.Background(), &taskv1.UpdateTaskCommentRequest{
		Id:   commentID.String(),
		Body: "Edited",
	}, &coredto.UserInfo{UserID: uuid.NewString()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func TestDeleteTaskCommentForbiddenForNonAuthor(t *testing.T) {
	commentID := uuid.New()
	authorID := uuid.New()
	otherUserID := uuid.New()

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		getTaskCommentFn: func(_ context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
			if id != commentID {
				t.Fatalf("unexpected comment id: %s", id)
			}
			return &db.TaskComment{
				ID:       commentID,
				TaskID:   uuid.New(),
				AuthorID: authorID,
				Body:     "Original",
			}, nil
		},
	})

	res, err := svc.DeleteTaskComment(context.Background(), &taskv1.DeleteTaskCommentRequest{
		Id: commentID.String(),
	}, &coredto.UserInfo{UserID: otherUserID.String()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskCommentAppError(t, err, apperror.CodeForbidden, "only the comment author can delete this comment")
}

func TestDeleteTaskCommentSuccess(t *testing.T) {
	commentID := uuid.New()
	authorID := uuid.New()

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		getTaskCommentFn: func(_ context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
			if id != commentID {
				t.Fatalf("unexpected comment id: %s", id)
			}
			return &db.TaskComment{
				ID:       commentID,
				TaskID:   uuid.New(),
				AuthorID: authorID,
				Body:     "Original",
			}, nil
		},
		deleteTaskCommentFn: func(_ context.Context, params db.DeleteTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
			if params.ID != commentID {
				t.Fatalf("unexpected comment id: %s", params.ID)
			}
			if !params.DeletedBy.Valid || params.DeletedBy.Bytes != authorID {
				t.Fatalf("unexpected deleted_by: %#v", params.DeletedBy)
			}
			return &db.TaskComment{ID: commentID}, nil
		},
	})

	res, err := svc.DeleteTaskComment(context.Background(), &taskv1.DeleteTaskCommentRequest{
		Id: commentID.String(),
	}, &coredto.UserInfo{UserID: authorID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("expected success response, got %#v", res)
	}
}

func TestListTaskCommentsSuccess(t *testing.T) {
	taskID := uuid.New()
	commentID := uuid.New()
	authorID := uuid.New()

	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		listTaskCommentsFn: func(_ context.Context, got uuid.UUID) ([]db.TaskComment, *apperror.AppError) {
			if got != taskID {
				t.Fatalf("unexpected task id: %s", got)
			}
			return []db.TaskComment{
				{
					ID:       commentID,
					TaskID:   taskID,
					AuthorID: authorID,
					Body:     "Looks good",
				},
			}, nil
		},
	})

	res, err := svc.ListTaskComments(context.Background(), &taskv1.ListTaskCommentsRequest{
		TaskId: taskID.String(),
	}, &coredto.UserInfo{UserID: authorID.String()})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(res.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(res.Comments))
	}
	if res.Comments[0].Id != commentID.String() {
		t.Fatalf("unexpected comment id: %s", res.Comments[0].Id)
	}
}

func TestListTaskCommentsValidationAndRepoError(t *testing.T) {
	authorID := uuid.New()
	svc := mustTaskCommentService(t, &stubTaskCommentRepository{
		listTaskCommentsFn: func(context.Context, uuid.UUID) ([]db.TaskComment, *apperror.AppError) {
			t.Fatal("repository should not be called for invalid task id")
			return nil, nil
		},
	})

	res, err := svc.ListTaskComments(context.Background(), &taskv1.ListTaskCommentsRequest{
		TaskId: "bad-uuid",
	}, &coredto.UserInfo{UserID: authorID.String()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	assertTaskCommentAppError(t, err, apperror.CodeValidation, "invalid uuid")

	taskID := uuid.New()
	repoErr := &apperror.AppError{Code: apperror.CodeInternal, Message: "failed to list task comments"}
	svc = mustTaskCommentService(t, &stubTaskCommentRepository{
		listTaskCommentsFn: func(_ context.Context, got uuid.UUID) ([]db.TaskComment, *apperror.AppError) {
			if got != taskID {
				t.Fatalf("unexpected task id: %s", got)
			}
			return nil, repoErr
		},
	})

	res, err = svc.ListTaskComments(context.Background(), &taskv1.ListTaskCommentsRequest{
		TaskId: taskID.String(),
	}, &coredto.UserInfo{UserID: authorID.String()})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if err != repoErr {
		t.Fatalf("expected repo error to be returned unchanged, got %#v", err)
	}
}

func mustTaskCommentService(t *testing.T, repo *stubTaskCommentRepository) TaskCommentService {
	t.Helper()

	svc, err := NewTaskCommentService(repo, servicetestutil.NewAllowAuthorizer())
	if err != nil {
		t.Fatalf("failed to construct task comment service: %v", err)
	}

	return svc
}

func assertTaskCommentAppError(t *testing.T, err *apperror.AppError, code apperror.ErrorCode, message string) {
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
