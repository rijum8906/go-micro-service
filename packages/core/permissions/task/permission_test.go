package permissions_test

import (
	"testing"

	task "github.com/rijum8906/relay/packages/core/permissions/task"
	"github.com/stretchr/testify/require"
)

func TestPermissionHelpersValidateRoles(t *testing.T) {
	require.True(t, task.IsSystemRole(string(task.RoleOwner)))
	require.True(t, task.IsSystemRole(string(task.RoleAdmin)))
	require.True(t, task.IsSystemRole(string(task.RoleMember)))
	require.False(t, task.IsSystemRole("reviewer"))

	require.True(t, task.IsValidCustomRole("reviewer"))
	require.False(t, task.IsValidCustomRole(""))
	require.False(t, task.IsValidCustomRole(string(task.RoleAdmin)))
}

func TestPermissionHelpersValidateResourcePermissions(t *testing.T) {
	require.True(t, task.IsValidResourcePermission(task.ResourceProject, task.PermissionCanAddMember))
	require.True(t, task.IsValidResourcePermission(task.ResourceTask, task.PermissionCanAssign))
	require.True(t, task.IsValidResourcePermission(task.ResourceTaskComment, task.PermissionCanDelete))

	require.False(t, task.IsValidResourcePermission(task.ResourceProject, task.PermissionCanComment))
	require.False(t, task.IsValidResourcePermission(task.ResourceTaskComment, task.PermissionCanAssign))
	require.False(t, task.IsValidResourcePermission("unknown", task.PermissionCanView))
}

func TestPermissionHelpersGenerateOpenFGAObjects(t *testing.T) {
	require.Equal(t, "role:project-1_reviewer", task.GenerateCustomRoleObject("project-1", "reviewer"))
	require.Equal(t, "permission:project-1_task_can_comment", task.GeneratePermissionObject("project-1", task.ResourceTask, task.PermissionCanComment))
}

func TestGetDefaultPermissionsForRole(t *testing.T) {
	ownerPermissions := task.GetDefaultPermissionsForRole(string(task.RoleOwner))
	adminPermissions := task.GetDefaultPermissionsForRole(string(task.RoleAdmin))
	memberPermissions := task.GetDefaultPermissionsForRole(string(task.RoleMember))

	require.ElementsMatch(t, task.AllPermissions, ownerPermissions)
	require.Contains(t, adminPermissions, task.PermissionCanManageTasks)
	require.Contains(t, adminPermissions, task.PermissionCanAssign)
	require.NotContains(t, adminPermissions, task.PermissionCanDelete)
	require.Contains(t, memberPermissions, task.PermissionCanComment)
	require.NotContains(t, memberPermissions, task.PermissionCanManage)
	require.Empty(t, task.GetDefaultPermissionsForRole("unknown"))
}
