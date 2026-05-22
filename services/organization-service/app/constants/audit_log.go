package constants

import "strings"

// Audit Log Actions - Organization & Membership
const (
	// Organization Actions
	AuditLogActionCreateOrganization  = "organization.create"
	AuditLogActionUpdateOrganization  = "organization.update"
	AuditLogActionDeleteOrganization  = "organization.delete"
	AuditLogActionArchiveOrganization = "organization.archive"
	AuditLogActionRestoreOrganization = "organization.restore"

	// Organization Membership Actions
	AuditLogActionInviteMember      = "membership.invite"
	AuditLogActionAcceptInvitation  = "membership.accept_invitation"
	AuditLogActionDeclineInvitation = "membership.decline_invitation"
	AuditLogActionRevokeInvitation  = "membership.revoke_invitation"
	AuditLogActionJoinOrganization  = "membership.join"        // User joined via invite
	AuditLogActionLeaveOrganization = "membership.leave"       // User voluntarily left
	AuditLogActionAddMember         = "membership.add"         // Admin added member
	AuditLogActionRemoveMember      = "membership.remove"      // Admin removed member
	AuditLogActionBanMember         = "membership.ban"         // Admin banned member
	AuditLogActionUnbanMember       = "membership.unban"       // Admin unbanned member
	AuditLogActionSuspendMember     = "membership.suspend"     // Admin suspended member
	AuditLogActionActivateMember    = "membership.activate"    // Admin activated suspended member
	AuditLogActionChangeMemberRole  = "membership.change_role" // Role changed
	AuditLogActionTransferOwnership = "membership.transfer_ownership"

	// Organization Team Actions
	AuditLogActionCreateTeam           = "team.create"
	AuditLogActionUpdateTeam           = "team.update"
	AuditLogActionDeleteTeam           = "team.delete"
	AuditLogActionArchiveTeam          = "team.archive"
	AuditLogActionRestoreTeam          = "team.restore"
	AuditLogActionAddTeamMember        = "team.add_member"
	AuditLogActionRemoveTeamMember     = "team.remove_member"
	AuditLogActionChangeTeamMemberRole = "team.change_member_role"

	// Permission Actions
	AuditLogActionGrantPermission       = "permission.grant"
	AuditLogActionRevokePermission      = "permission.revoke"
	AuditLogActionUpdateRolePermissions = "permission.update_role"

	// Settings Actions
	AuditLogActionUpdateOrganizationSettings = "settings.organization.update"
	AuditLogActionUpdateSecuritySettings     = "settings.security.update"
	AuditLogActionUpdateBillingSettings      = "settings.billing.update"
)

// Action Categories for grouping
const (
	AuditCategoryOrganization = "organization"
	AuditCategoryMembership   = "membership"
	AuditCategoryTeam         = "team"
	AuditCategoryPermission   = "permission"
	AuditCategorySettings     = "settings"
)

// GetActionCategory returns the category of an audit action
func GetActionCategory(action string) string {
	switch {
	case strings.HasPrefix(action, "organization."):
		return AuditCategoryOrganization
	case strings.HasPrefix(action, "membership."):
		return AuditCategoryMembership
	case strings.HasPrefix(action, "team."):
		return AuditCategoryTeam
	case strings.HasPrefix(action, "permission."):
		return AuditCategoryPermission
	case strings.HasPrefix(action, "settings."):
		return AuditCategorySettings
	default:
		return "unknown"
	}
}

// Audit Severity Levels
const (
	AuditSeverityLow      = "low"      // Informational, no security impact
	AuditSeverityMedium   = "medium"   // Important but not security critical
	AuditSeverityHigh     = "high"     // Security relevant (role changes, permissions)
	AuditSeverityCritical = "critical" // Security breach, ownership changes
)

// GetActionSeverity returns the severity level of an audit action
func GetActionSeverity(action string) string {
	highSeverityActions := map[string]bool{
		AuditLogActionTransferOwnership:  true,
		AuditLogActionGrantPermission:    true,
		AuditLogActionRevokePermission:   true,
		AuditLogActionDeleteOrganization: true,
		AuditLogActionBanMember:          true,
		AuditLogActionRemoveMember:       true,
	}

	if highSeverityActions[action] {
		return AuditSeverityHigh
	}

	mediumSeverityActions := map[string]bool{
		AuditLogActionChangeMemberRole:    true,
		AuditLogActionCreateOrganization:  true,
		AuditLogActionUpdateOrganization:  true,
		AuditLogActionArchiveOrganization: true,
		AuditLogActionInviteMember:        true,
		AuditLogActionSuspendMember:       true,
		AuditLogActionCreateTeam:          true,
		AuditLogActionUpdateTeam:          true,
	}

	if mediumSeverityActions[action] {
		return AuditSeverityMedium
	}

	return AuditSeverityLow
}
