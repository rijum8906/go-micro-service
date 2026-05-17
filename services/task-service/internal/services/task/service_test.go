package task

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	taskpermissions "github.com/rijum8906/relay/packages/core/permissions/task"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	servicetestutil "github.com/rijum8906/relay/services/task-service/internal/services/testutil"
	"google.golang.org/grpc/codes"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type stubQuerier struct {
	db.Querier

	createProjectFn           func(context.Context, db.CreateProjectParams) (db.Project, error)
	getActiveMembershipFn     func(context.Context, db.GetActiveProjectMembershipParams) (db.ProjectMembership, error)
	addProjectMemberFn        func(context.Context, db.AddProjectMemberParams) (db.ProjectMembership, error)
	updateProjectMemberRoleFn func(context.Context, db.UpdateProjectMemberRoleParams) (db.ProjectMembership, error)
	createTaskFn              func(context.Context, db.CreateTaskParams) (db.Task, error)
	getActiveTaskAssignmentFn func(context.Context, db.GetActiveTaskAssignmentParams) (db.TaskAssignment, error)
	assignTaskFn              func(context.Context, db.AssignTaskParams) (db.TaskAssignment, error)
	createTaskCommentFn       func(context.Context, db.CreateTaskCommentParams) (db.TaskComment, error)
	getTaskCommentFn          func(context.Context, uuid.UUID) (db.TaskComment, error)
	deleteTaskCommentFn       func(context.Context, db.DeleteTaskCommentParams) (db.TaskComment, error)
}

func (s stubQuerier) CreateProject(ctx context.Context, params db.CreateProjectParams) (db.Project, error) {
	if s.createProjectFn == nil {
		panic("unexpected CreateProject call")
	}
	return s.createProjectFn(ctx, params)
}

func (s stubQuerier) GetActiveProjectMembership(ctx context.Context, params db.GetActiveProjectMembershipParams) (db.ProjectMembership, error) {
	if s.getActiveMembershipFn == nil {
		panic("unexpected GetActiveProjectMembership call")
	}
	return s.getActiveMembershipFn(ctx, params)
}

func (s stubQuerier) AddProjectMember(ctx context.Context, params db.AddProjectMemberParams) (db.ProjectMembership, error) {
	if s.addProjectMemberFn == nil {
		panic("unexpected AddProjectMember call")
	}
	return s.addProjectMemberFn(ctx, params)
}

func (s stubQuerier) UpdateProjectMemberRole(ctx context.Context, params db.UpdateProjectMemberRoleParams) (db.ProjectMembership, error) {
	if s.updateProjectMemberRoleFn == nil {
		panic("unexpected UpdateProjectMemberRole call")
	}
	return s.updateProjectMemberRoleFn(ctx, params)
}

func (s stubQuerier) CreateTask(ctx context.Context, params db.CreateTaskParams) (db.Task, error) {
	if s.createTaskFn == nil {
		panic("unexpected CreateTask call")
	}
	return s.createTaskFn(ctx, params)
}

func (s stubQuerier) GetActiveTaskAssignment(ctx context.Context, params db.GetActiveTaskAssignmentParams) (db.TaskAssignment, error) {
	if s.getActiveTaskAssignmentFn == nil {
		panic("unexpected GetActiveTaskAssignment call")
	}
	return s.getActiveTaskAssignmentFn(ctx, params)
}

func (s stubQuerier) AssignTask(ctx context.Context, params db.AssignTaskParams) (db.TaskAssignment, error) {
	if s.assignTaskFn == nil {
		panic("unexpected AssignTask call")
	}
	return s.assignTaskFn(ctx, params)
}

func (s stubQuerier) CreateTaskComment(ctx context.Context, params db.CreateTaskCommentParams) (db.TaskComment, error) {
	if s.createTaskCommentFn == nil {
		panic("unexpected CreateTaskComment call")
	}
	return s.createTaskCommentFn(ctx, params)
}

func (s stubQuerier) GetTaskComment(ctx context.Context, id uuid.UUID) (db.TaskComment, error) {
	if s.getTaskCommentFn == nil {
		panic("unexpected GetTaskComment call")
	}
	return s.getTaskCommentFn(ctx, id)
}

func (s stubQuerier) DeleteTaskComment(ctx context.Context, params db.DeleteTaskCommentParams) (db.TaskComment, error) {
	if s.deleteTaskCommentFn == nil {
		panic("unexpected DeleteTaskComment call")
	}
	return s.deleteTaskCommentFn(ctx, params)
}

func TestNewRequiresQueries(t *testing.T) {
	svc, appErr := New(nil, nil)
	if appErr == nil {
		t.Fatal("expected constructor error")
	}
	if svc != nil {
		t.Fatal("expected nil service")
	}
}

func TestCreateProjectRequiresIncomingUserMetadata(t *testing.T) {
	svc := newTestService(t, stubQuerier{
		createProjectFn: func(context.Context, db.CreateProjectParams) (db.Project, error) {
			t.Fatal("query should not be called")
			return db.Project{}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), nil)

	res, err := svc.CreateProject(context.Background(), &taskv1.CreateProjectRequest{Name: "Platform"})
	if res != nil {
		t.Fatalf("expected nil response, got %#v", res)
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}

func TestCreateProjectWritesOwnerTuple(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tuples := &servicetestutil.TupleManager{}

	svc := newTestService(t, stubQuerier{
		createProjectFn: func(_ context.Context, params db.CreateProjectParams) (db.Project, error) {
			return db.Project{
				ID:        projectID,
				CreatedBy: params.CreatedBy,
				Name:      params.Name,
				Status:    "active",
			}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), tuples)

	_, err := svc.CreateProject(contextWithUser(userID), &taskv1.CreateProjectRequest{Name: "Platform"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !tuples.HasWrite(authz.ProjectRoleTuple(projectID, string(taskpermissions.RoleOwner), userID)) {
		t.Fatalf("missing project owner tuple: %#v", tuples.Writes)
	}
}

func TestAddProjectMemberWritesRoleTuple(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	tuples := &servicetestutil.TupleManager{}

	svc := newTestService(t, stubQuerier{
		getActiveMembershipFn: func(context.Context, db.GetActiveProjectMembershipParams) (db.ProjectMembership, error) {
			return db.ProjectMembership{}, pgx.ErrNoRows
		},
		addProjectMemberFn: func(_ context.Context, params db.AddProjectMemberParams) (db.ProjectMembership, error) {
			return db.ProjectMembership{
				ID:        uuid.New(),
				ProjectID: params.ProjectID,
				UserID:    params.UserID,
				Role:      params.Role,
			}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), tuples)

	_, err := svc.AddProjectMember(contextWithUser(uuid.New()), &taskv1.AddProjectMemberRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      string(taskpermissions.RoleAdmin),
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !tuples.HasWrite(authz.ProjectRoleTuple(projectID, string(taskpermissions.RoleAdmin), userID)) {
		t.Fatalf("missing project member tuple: %#v", tuples.Writes)
	}
}

func TestUpdateProjectMemberRoleReplacesTuple(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	tuples := &servicetestutil.TupleManager{}
	lookupCount := 0

	svc := newTestService(t, stubQuerier{
		getActiveMembershipFn: func(_ context.Context, params db.GetActiveProjectMembershipParams) (db.ProjectMembership, error) {
			lookupCount++
			return db.ProjectMembership{
				ID:        uuid.New(),
				ProjectID: params.ProjectID,
				UserID:    params.UserID,
				Role:      string(taskpermissions.RoleMember),
			}, nil
		},
		updateProjectMemberRoleFn: func(_ context.Context, params db.UpdateProjectMemberRoleParams) (db.ProjectMembership, error) {
			return db.ProjectMembership{
				ID:        uuid.New(),
				ProjectID: params.ProjectID,
				UserID:    params.UserID,
				Role:      params.Role,
			}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), tuples)

	_, err := svc.UpdateProjectMemberRole(contextWithUser(uuid.New()), &taskv1.UpdateProjectMemberRoleRequest{
		ProjectId: projectID.String(),
		UserId:    userID.String(),
		Role:      string(taskpermissions.RoleAdmin),
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if lookupCount != 1 {
		t.Fatalf("expected current membership lookup, got %d", lookupCount)
	}
	if !tuples.HasDelete(authz.DeleteTuple(authz.ProjectRoleTuple(projectID, string(taskpermissions.RoleMember), userID))) {
		t.Fatalf("missing old role delete: %#v", tuples.Deletes)
	}
	if !tuples.HasWrite(authz.ProjectRoleTuple(projectID, string(taskpermissions.RoleAdmin), userID)) {
		t.Fatalf("missing new role write: %#v", tuples.Writes)
	}
}

func TestCreateTaskWritesCreatorAndProjectTuples(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	projectID := uuid.New()
	tuples := &servicetestutil.TupleManager{}

	svc := newTestService(t, stubQuerier{
		createTaskFn: func(_ context.Context, params db.CreateTaskParams) (db.Task, error) {
			if !params.ProjectID.Valid || params.ProjectID.Bytes != projectID {
				t.Fatalf("unexpected project id: %#v", params.ProjectID)
			}
			return db.Task{
				ID:        taskID,
				ProjectID: params.ProjectID,
				CreatedBy: params.CreatedBy,
				Title:     params.Title,
				Priority:  params.Priority,
			}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), tuples)

	_, err := svc.CreateTask(contextWithUser(userID), &taskv1.CreateTaskRequest{
		ProjectId: projectID.String(),
		Title:     "Ship task-service",
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !tuples.HasWrite(authz.TaskCreatorTuple(taskID, userID)) {
		t.Fatalf("missing task creator tuple: %#v", tuples.Writes)
	}
	if !tuples.HasWrite(authz.TaskProjectTuple(taskID, projectID)) {
		t.Fatalf("missing task project tuple: %#v", tuples.Writes)
	}
}

func TestAssignTaskWritesAssigneeTuple(t *testing.T) {
	taskID := uuid.New()
	assigneeID := uuid.New()
	tuples := &servicetestutil.TupleManager{}

	svc := newTestService(t, stubQuerier{
		getActiveTaskAssignmentFn: func(context.Context, db.GetActiveTaskAssignmentParams) (db.TaskAssignment, error) {
			return db.TaskAssignment{}, pgx.ErrNoRows
		},
		assignTaskFn: func(_ context.Context, params db.AssignTaskParams) (db.TaskAssignment, error) {
			return db.TaskAssignment{
				ID:           uuid.New(),
				TaskID:       params.TaskID,
				AssigneeType: params.AssigneeType,
				AssigneeID:   params.AssigneeID,
				AssignedBy:   params.AssignedBy,
			}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), tuples)

	_, err := svc.AssignTask(contextWithUser(uuid.New()), &taskv1.AssignTaskRequest{
		TaskId:       taskID.String(),
		AssigneeType: "user",
		AssigneeId:   assigneeID.String(),
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !tuples.HasWrite(authz.TaskAssigneeTuple(taskID, "user", assigneeID)) {
		t.Fatalf("missing assignee tuple: %#v", tuples.Writes)
	}
}

func TestCreateAndDeleteTaskCommentSyncTuples(t *testing.T) {
	taskID := uuid.New()
	commentID := uuid.New()
	authorID := uuid.New()
	tuples := &servicetestutil.TupleManager{}

	svc := newTestService(t, stubQuerier{
		createTaskCommentFn: func(_ context.Context, params db.CreateTaskCommentParams) (db.TaskComment, error) {
			return db.TaskComment{
				ID:       commentID,
				TaskID:   params.TaskID,
				AuthorID: params.AuthorID,
				Body:     params.Body,
			}, nil
		},
		getTaskCommentFn: func(_ context.Context, id uuid.UUID) (db.TaskComment, error) {
			return db.TaskComment{
				ID:       id,
				TaskID:   taskID,
				AuthorID: authorID,
				Body:     "Looks good",
			}, nil
		},
		deleteTaskCommentFn: func(_ context.Context, params db.DeleteTaskCommentParams) (db.TaskComment, error) {
			return db.TaskComment{
				ID:       params.ID,
				TaskID:   taskID,
				AuthorID: authorID,
				Body:     "Looks good",
			}, nil
		},
	}, servicetestutil.NewAllowAuthorizer(), tuples)

	_, err := svc.CreateTaskComment(contextWithUser(authorID), &taskv1.CreateTaskCommentRequest{
		TaskId: taskID.String(),
		Body:   "Looks good",
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !tuples.HasWrite(authz.CommentAuthorTuple(commentID, authorID)) || !tuples.HasWrite(authz.CommentTaskTuple(commentID, taskID)) {
		t.Fatalf("missing comment tuple writes: %#v", tuples.Writes)
	}

	_, err = svc.DeleteTaskComment(contextWithUser(authorID), &taskv1.DeleteTaskCommentRequest{Id: commentID.String()})
	if err != nil {
		t.Fatalf("expected delete success, got error: %v", err)
	}
	if !tuples.HasDelete(authz.DeleteTuple(authz.CommentAuthorTuple(commentID, authorID))) {
		t.Fatalf("missing comment author delete: %#v", tuples.Deletes)
	}
	if !tuples.HasDelete(authz.DeleteTuple(authz.CommentTaskTuple(commentID, taskID))) {
		t.Fatalf("missing comment task delete: %#v", tuples.Deletes)
	}
}

func newTestService(t *testing.T, q db.Querier, authorizer authz.Authorizer, tuples *servicetestutil.TupleManager) *service {
	t.Helper()

	svc, appErr := newService(q, authorizer, tuples)
	if appErr != nil {
		t.Fatalf("failed to construct service: %v", appErr)
	}

	return svc
}

func contextWithUser(userID uuid.UUID) context.Context {
	return grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs(
		coredto.MetaUserIDKey, userID.String(),
	))
}

var _ db.Querier = stubQuerier{}
