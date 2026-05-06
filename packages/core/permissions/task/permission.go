// Package permissions contains the permission and role constants for task authorization.
package permissions

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

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
	PermissionCanEdit             = "can_edit"
	PermissionCanUpdateStatus     = "can_update_status"
	PermissionCanUpdateProgress   = "can_update_progress"
	PermissionCanComment          = "can_comment"
	PermissionCanManage           = "can_manage"
	PermissionCanAssign           = "can_assign"
)
