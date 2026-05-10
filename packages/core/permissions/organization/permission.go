// Package permissions provides organization and team level permissions for OpenFGA
package permissions

import "slices"

// Resource types to help with OpenFGA tuple construction
const (
	ResourceOrganization = "organization"
	ResourceTeam         = "team"
	ResourceUser         = "user"
	ResourceRole         = "role"
	ResourcePermission   = "permission"
)

// ============================================================================
// Organization Level Permissions
// ============================================================================

// Organization basic permissions
const (
	PermissionCanEditOrganization   = "can_edit_organization"
	PermissionCanDeleteOrganization = "can_delete_organization"
	PermissionCanViewOrganization   = "can_view_organization"
	PermissionCanTransferOwnership  = "can_transfer_ownership"
)

// Member Management Permissions
const (
	PermissionCanAddMember          = "can_add_member"
	PermissionCanRemoveMember       = "can_remove_member"
	PermissionCanViewMembers        = "can_view_members"
	PermissionCanChangeMemberRole   = "can_change_member_role"
	PermissionCanChangeMemberStatus = "can_change_member_status"
	PermissionCanInviteMember       = "can_invite_member"
	PermissionCanApproveJoinRequest = "can_approve_join_request"
)

// Team Management Permissions
const (
	PermissionCanCreateTeam = "can_create_team"
	PermissionCanViewTeams  = "can_view_teams"
)

// Role Management Permissions
const (
	PermissionCanCreateRole = "can_create_role"
	PermissionCanDeleteRole = "can_delete_role"
	PermissionCanViewRoles  = "can_view_roles"
	PermissionCanEditRole   = "can_edit_role"
	PermissionCanAssignRole = "can_assign_role"
)

// ============================================================================
// Team Level Permissions
// ============================================================================

// Team basic permissions
const (
	PermissionCanEditTeam   = "can_edit_team"
	PermissionCanDeleteTeam = "can_delete_team"
	PermissionCanViewTeam   = "can_view_team"
)

// Team Member Management Permissions
const (
	PermissionCanAddTeamMember    = "can_add_team_member"
	PermissionCanRemoveTeamMember = "can_remove_team_member"
	PermissionCanViewTeamMembers  = "can_view_team_members"
	PermissionCanInviteTeamMember = "can_invite_team_member"
)

// ============================================================================
// System Roles
// ============================================================================

// System roles — these are built-in and cannot be deleted
const (
	RoleOwner   = "owner"
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleLeader  = "leader"
	RoleMember  = "member"
	RoleViewer  = "viewer"
	RoleGuest   = "guest"
)

// SystemRoles is the ordered list from highest to lowest privilege.
// Used for role hierarchy comparisons.
var SystemRoles = []string{
	RoleOwner,
	RoleAdmin,
	RoleManager,
	RoleLeader,
	RoleMember,
	RoleViewer,
	RoleGuest,
}

// roleRank maps a role to its privilege level (lower = higher privilege)
var roleRank = func() map[string]int {
	m := make(map[string]int, len(SystemRoles))
	for i, r := range SystemRoles {
		m[r] = i
	}
	return m
}()

// ============================================================================
// Permission Collections
// ============================================================================

// OrgPermissions is the full set of permissions that can be granted
// to a custom role at the organization level.
var OrgPermissions = []string{
	// Organization basic
	PermissionCanEditOrganization,
	PermissionCanDeleteOrganization,
	PermissionCanViewOrganization,
	PermissionCanTransferOwnership,

	// Member management
	PermissionCanAddMember,
	PermissionCanRemoveMember,
	PermissionCanViewMembers,
	PermissionCanChangeMemberRole,
	PermissionCanChangeMemberStatus,
	PermissionCanInviteMember,
	PermissionCanApproveJoinRequest,

	// Team management
	PermissionCanCreateTeam,
	PermissionCanDeleteTeam,
	PermissionCanViewTeams,

	// Role management
	PermissionCanCreateRole,
	PermissionCanDeleteRole,
	PermissionCanViewRoles,
	PermissionCanEditRole,
	PermissionCanAssignRole,
}

// TeamPermissions is the full set of permissions that can be granted
// to a custom role at the team level.
var TeamPermissions = []string{
	// Team basic
	PermissionCanEditTeam,
	PermissionCanDeleteTeam,
	PermissionCanViewTeam,

	// Team member management
	PermissionCanAddTeamMember,
	PermissionCanRemoveTeamMember,
	PermissionCanViewTeamMembers,
	PermissionCanInviteTeamMember,
}

// AllPermissions combines all possible permissions
var AllPermissions = append(OrgPermissions, TeamPermissions...)

// validRoles creates a set of valid system roles for quick lookup
var validRoles = func() map[string]struct{} {
	m := make(map[string]struct{}, len(SystemRoles))
	for _, r := range SystemRoles {
		m[r] = struct{}{}
	}
	return m
}()

// validPermissions creates a set of valid permissions for quick lookup
var validPermissions = func() map[string]struct{} {
	m := make(map[string]struct{}, len(AllPermissions))
	for _, p := range AllPermissions {
		m[p] = struct{}{}
	}
	return m
}()

// ============================================================================
// Helper Functions
// ============================================================================

// IsValidRole returns true if the role is a known system role.
func IsValidRole(role string) bool {
	_, ok := validRoles[role]
	return ok
}

// IsValidPermission returns true if the permission is a known permission.
func IsValidPermission(permission string) bool {
	_, ok := validPermissions[permission]
	return ok
}

// IsOrgPermission returns true if the permission is an organization-level permission.
func IsOrgPermission(permission string) bool {
	return slices.Contains(OrgPermissions, permission)
}

// IsTeamPermission returns true if the permission is a team-level permission.
func IsTeamPermission(permission string) bool {
	return slices.Contains(TeamPermissions, permission)
}

// IsSystemRole returns true for built-in roles that cannot be deleted or modified.
func IsSystemRole(role string) bool {
	return IsValidRole(role)
}

// RoleRank returns the privilege rank of a role (0 = highest privilege).
// Returns -1 for unknown roles.
func RoleRank(role string) int {
	if rank, ok := roleRank[role]; ok {
		return rank
	}
	return -1
}

// CanActorManageTarget returns true if actorRole has higher privilege than targetRole.
// Owners can manage everyone. Actors cannot manage peers or superiors.
func CanActorManageTarget(actorRole, targetRole string) bool {
	actorRank := RoleRank(actorRole)
	targetRank := RoleRank(targetRole)
	if actorRank == -1 || targetRank == -1 {
		return false
	}
	return actorRank < targetRank // lower rank index = higher privilege
}

// ============================================================================
// Object ID Builders
// ============================================================================

// PermissionObjectID returns the OpenFGA object ID for a permission
// scoped to an organization, e.g. "org123_can_edit_organization"
func PermissionObjectID(orgID, permission string) string {
	return orgID + "_" + permission
}

// PermissionObject returns the full OpenFGA object string for a permission,
// e.g. "permission:org123_can_edit_organization"
func PermissionObject(orgID, permission string) string {
	return ResourcePermission + ":" + PermissionObjectID(orgID, permission)
}

// TeamPermissionObjectID returns the OpenFGA object ID for a permission
// scoped to a team, e.g. "team123_can_edit_team"
func TeamPermissionObjectID(teamID, permission string) string {
	return teamID + "_" + permission
}

// TeamPermissionObject returns the full OpenFGA object string for a team permission,
// e.g. "permission:team123_can_edit_team"
func TeamPermissionObject(teamID, permission string) string {
	return ResourcePermission + ":" + TeamPermissionObjectID(teamID, permission)
}

// OrganizationObject returns the full OpenFGA object string for an organization
func OrganizationObject(orgID string) string {
	return ResourceOrganization + ":" + orgID
}

// TeamObject returns the full OpenFGA object string for a team
func TeamObject(teamID string) string {
	return ResourceTeam + ":" + teamID
}

// UserObject returns the full OpenFGA object string for a user
func UserObject(userID string) string {
	return ResourceUser + ":" + userID
}

// RoleObject returns the full OpenFGA object string for a role
func RoleObject(roleID string) string {
	return ResourceRole + ":" + roleID
}

// ============================================================================
// Role to Default Permissions Mapping
// ============================================================================

// GetDefaultPermissionsForRole returns the default permissions for a system role
func GetDefaultPermissionsForRole(role string, orgID string) []string {
	switch role {
	case RoleOwner:
		return OrgPermissions // Owners get all org permissions
	case RoleAdmin:
		return []string{
			PermissionCanViewOrganization,
			PermissionCanAddMember,
			PermissionCanRemoveMember,
			PermissionCanViewMembers,
			PermissionCanChangeMemberRole,
			PermissionCanChangeMemberStatus,
			PermissionCanInviteMember,
			PermissionCanApproveJoinRequest,
			PermissionCanCreateTeam,
			PermissionCanDeleteTeam,
			PermissionCanViewTeams,
			PermissionCanCreateRole,
			PermissionCanDeleteRole,
			PermissionCanViewRoles,
			PermissionCanEditRole,
			PermissionCanAssignRole,
		}
	case RoleManager:
		return []string{
			PermissionCanViewOrganization,
			PermissionCanViewMembers,
			PermissionCanInviteMember,
			PermissionCanViewTeams,
			PermissionCanViewRoles,
		}
	case RoleLeader:
		return []string{
			PermissionCanViewOrganization,
			PermissionCanViewMembers,
			PermissionCanViewTeams,
		}
	case RoleMember:
		return []string{
			PermissionCanViewOrganization,
			PermissionCanViewMembers,
		}
	case RoleViewer:
		return []string{
			PermissionCanViewOrganization,
		}
	case RoleGuest:
		return []string{} // Guests have no default permissions
	default:
		return []string{}
	}
}

// ============================================================================
// Permission Validation
// ============================================================================

// ValidatePermissionAssignment checks if a permission can be assigned to a role
func ValidatePermissionAssignment(permission string, role string) bool {
	// Owner can have any permission
	if role == RoleOwner {
		return true
	}

	// Certain permissions are owner-only
	ownerOnlyPermissions := []string{
		PermissionCanDeleteOrganization,
		PermissionCanTransferOwnership,
	}

	for _, ownerOnly := range ownerOnlyPermissions {
		if permission == ownerOnly && role != RoleOwner {
			return false
		}
	}

	return IsValidPermission(permission)
}
