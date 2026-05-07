// Package permissions contains the permission for organization
package permissions

const (
	PermissionCanEdit         = "can_edit"
	PermissionCanDelete       = "can_delete"
	PermissionCanAddMember    = "can_add_member"
	PermissionCanRemoveMember = "can_remove_member"
	PermissionCanViewMember   = "can_view_member"
	PermissionCanCreateTeam   = "can_create_team"
)

const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleOwner  = "owner"
	RoleMember = "member"
	RoleLeader = "leader"
)

func ValidateRole(role string) bool {
	switch role {
	case RoleAdmin, RoleUser, RoleOwner, RoleMember, RoleLeader:
		return true
	default:
		return false
	}
}
