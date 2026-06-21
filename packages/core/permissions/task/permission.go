// Package permissions contains the permission and role constants for task authorization.
package permissions

import "slices"

const (
	ResourceProject     = "project"
	ResourceTask        = "task"
	ResourceTaskComment = "task_comment"
	ResourceRole        = "role"
	ResourcePermission  = "permission"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

var SystemRoles = []string{
	string(RoleOwner),
	string(RoleAdmin),
	string(RoleMember),
}

const (
	PermissionCanView             = "can_view"
	PermissionCanContributeTasks  = "can_contribute_tasks"
	PermissionCanManageTasks      = "can_manage_tasks"
	PermissionCanUpdate           = "can_update"
	PermissionCanComplete         = "can_complete"
	PermissionCanArchive          = "can_archive"
	PermissionCanDelete           = "can_delete"
	PermissionCanAddMember        = "can_add_member"
	PermissionCanRemoveMember     = "can_remove_member"
	PermissionCanChangeMemberRole = "can_change_member_role"

	PermissionCanEdit           = "can_edit"
	PermissionCanUpdateStatus   = "can_update_status"
	PermissionCanUpdateProgress = "can_update_progress"
	PermissionCanComment        = "can_comment"
	PermissionCanManage         = "can_manage"
	PermissionCanAssign         = "can_assign"
)

var ProjectPermissions = []string{
	PermissionCanView,
	PermissionCanContributeTasks,
	PermissionCanManageTasks,
	PermissionCanUpdate,
	PermissionCanComplete,
	PermissionCanArchive,
	PermissionCanDelete,
	PermissionCanAddMember,
	PermissionCanRemoveMember,
	PermissionCanChangeMemberRole,
}

var TaskPermissions = []string{
	PermissionCanView,
	PermissionCanEdit,
	PermissionCanUpdateStatus,
	PermissionCanUpdateProgress,
	PermissionCanComment,
	PermissionCanManage,
	PermissionCanAssign,
	PermissionCanArchive,
	PermissionCanDelete,
}

var TaskCommentPermissions = []string{
	PermissionCanView,
	PermissionCanEdit,
	PermissionCanDelete,
}

var AllPermissions = append(append(ProjectPermissions, TaskPermissions...), TaskCommentPermissions...)

func IsSystemRole(role string) bool {
	return slices.Contains(SystemRoles, role)
}

func IsValidPermission(permission string) bool {
	return slices.Contains(AllPermissions, permission)
}

func IsProjectPermission(permission string) bool {
	return slices.Contains(ProjectPermissions, permission)
}

func IsTaskPermission(permission string) bool {
	return slices.Contains(TaskPermissions, permission)
}

func IsTaskCommentPermission(permission string) bool {
	return slices.Contains(TaskCommentPermissions, permission)
}

func IsValidResource(resource string) bool {
	switch resource {
	case ResourceProject, ResourceTask, ResourceTaskComment:
		return true
	default:
		return false
	}
}

func IsValidCustomRole(role string) bool {
	return role != "" && !IsSystemRole(role)
}

func GenerateCustomRoleID(projectID, role string) string {
	return projectID + "_" + role
}

func GenerateCustomRoleObject(projectID, role string) string {
	return ResourceRole + ":" + GenerateCustomRoleID(projectID, role)
}

func GeneratePermissionID(projectID, resource, permission string) string {
	return projectID + "_" + resource + "_" + permission
}

func GeneratePermissionObject(projectID, resource, permission string) string {
	return ResourcePermission + ":" + GeneratePermissionID(projectID, resource, permission)
}

func GetDefaultPermissionsForRole(role string) []string {
	switch role {
	case string(RoleOwner):
		return AllPermissions
	case string(RoleAdmin):
		return []string{
			PermissionCanView,
			PermissionCanContributeTasks,
			PermissionCanManageTasks,
			PermissionCanUpdate,
			PermissionCanComplete,
			PermissionCanArchive,
			PermissionCanAddMember,
			PermissionCanRemoveMember,
			PermissionCanChangeMemberRole,
			PermissionCanEdit,
			PermissionCanUpdateStatus,
			PermissionCanUpdateProgress,
			PermissionCanComment,
			PermissionCanManage,
			PermissionCanAssign,
		}
	case string(RoleMember):
		return []string{
			PermissionCanView,
			PermissionCanContributeTasks,
			PermissionCanEdit,
			PermissionCanUpdateStatus,
			PermissionCanUpdateProgress,
			PermissionCanComment,
		}
	default:
		return []string{}
	}
}

type PermissionGrant struct {
	Resource   string
	Permission string
}

func ProjectPermission(permission string) PermissionGrant {
	return PermissionGrant{Resource: ResourceProject, Permission: permission}
}

func TaskPermission(permission string) PermissionGrant {
	return PermissionGrant{Resource: ResourceTask, Permission: permission}
}

func TaskCommentPermission(permission string) PermissionGrant {
	return PermissionGrant{Resource: ResourceTaskComment, Permission: permission}
}

func IsValidResourcePermission(resource, permission string) bool {
	switch resource {
	case ResourceProject:
		return IsProjectPermission(permission)
	case ResourceTask:
		return IsTaskPermission(permission)
	case ResourceTaskComment:
		return IsTaskCommentPermission(permission)
	default:
		return false
	}
}
