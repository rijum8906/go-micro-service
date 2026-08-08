package permissions_test

import (
	"context"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	task "github.com/rijum8906/relay/packages/core/permissions/task"
	"github.com/stretchr/testify/require"
)

type fakeTupleManager struct {
	writes  [][]client.ClientTupleKey
	deletes [][]client.ClientTupleKeyWithoutCondition
}

func (f *fakeTupleManager) Write(_ context.Context, writes []client.ClientTupleKey) *apperror.AppError {
	f.writes = append(f.writes, writes)
	return nil
}

func (f *fakeTupleManager) Read(_ context.Context, _ client.ClientReadRequest) (*client.ClientReadResponse, *apperror.AppError) {
	return nil, nil
}

func (f *fakeTupleManager) Check(_ context.Context, _ client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
	return nil, nil
}

func (f *fakeTupleManager) Delete(_ context.Context, deletes []client.ClientTupleKeyWithoutCondition) *apperror.AppError {
	f.deletes = append(f.deletes, deletes)
	return nil
}

func TestPermissionManagerCreateCustomRoleWritesTuples(t *testing.T) {
	tupleManager := &fakeTupleManager{}
	permissionManager := task.NewPermissionManagerWithTupleManager(tupleManager)

	appErr := permissionManager.CreateCustomRole(
		context.Background(),
		"user-1",
		"project-1",
		"reviewer",
		task.ProjectPermission(task.PermissionCanView),
		task.TaskPermission(task.PermissionCanComment),
	)

	require.Nil(t, appErr)
	require.Len(t, tupleManager.writes, 1)
	require.ElementsMatch(t, []client.ClientTupleKey{
		{
			User:     "project:project-1",
			Relation: task.ResourceProject,
			Object:   "role:project-1_reviewer",
		},
		{
			User:     "user:user-1",
			Relation: "assignee",
			Object:   "role:project-1_reviewer",
		},
		{
			User:     "role:project-1_reviewer",
			Relation: "granted_to",
			Object:   "permission:project-1_project_can_view",
		},
		{
			User:     "role:project-1_reviewer",
			Relation: "granted_to",
			Object:   "permission:project-1_task_can_comment",
		},
	}, tupleManager.writes[0])
}

func TestPermissionManagerDeleteCustomRoleDeletesTuples(t *testing.T) {
	tupleManager := &fakeTupleManager{}
	permissionManager := task.NewPermissionManagerWithTupleManager(tupleManager)

	appErr := permissionManager.DeleteCustomRole(
		context.Background(),
		"user-1",
		"project-1",
		"reviewer",
		task.ProjectPermission(task.PermissionCanView),
		task.TaskCommentPermission(task.PermissionCanDelete),
	)

	require.Nil(t, appErr)
	require.Len(t, tupleManager.deletes, 1)
	require.ElementsMatch(t, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "project:project-1",
			Relation: task.ResourceProject,
			Object:   "role:project-1_reviewer",
		},
		{
			User:     "user:user-1",
			Relation: "assignee",
			Object:   "role:project-1_reviewer",
		},
		{
			User:     "role:project-1_reviewer",
			Relation: "granted_to",
			Object:   "permission:project-1_project_can_view",
		},
		{
			User:     "role:project-1_reviewer",
			Relation: "granted_to",
			Object:   "permission:project-1_task_comment_can_delete",
		},
	}, tupleManager.deletes[0])
}

func TestPermissionManagerCreateCustomRoleRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		projectID string
		role      string
		grants    []task.PermissionGrant
	}{
		{
			name:      "missing user",
			userID:    "",
			projectID: "project-1",
			role:      "reviewer",
			grants:    []task.PermissionGrant{task.ProjectPermission(task.PermissionCanView)},
		},
		{
			name:      "missing project",
			userID:    "user-1",
			projectID: "",
			role:      "reviewer",
			grants:    []task.PermissionGrant{task.ProjectPermission(task.PermissionCanView)},
		},
		{
			name:      "system role",
			userID:    "user-1",
			projectID: "project-1",
			role:      string(task.RoleAdmin),
			grants:    []task.PermissionGrant{task.ProjectPermission(task.PermissionCanView)},
		},
		{
			name:      "invalid resource permission",
			userID:    "user-1",
			projectID: "project-1",
			role:      "reviewer",
			grants: []task.PermissionGrant{
				task.TaskCommentPermission(task.PermissionCanAssign),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tupleManager := &fakeTupleManager{}
			permissionManager := task.NewPermissionManagerWithTupleManager(tupleManager)

			appErr := permissionManager.CreateCustomRole(
				context.Background(),
				tt.userID,
				tt.projectID,
				tt.role,
				tt.grants...,
			)

			require.NotNil(t, appErr)
			require.Equal(t, apperror.CodeValidation, appErr.Code)
			require.Empty(t, tupleManager.writes)
		})
	}
}
