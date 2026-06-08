package jobs

const (
	JobUserAuthWildcard               = "user.auth.*.v1"
	JobUserRequestedPasswordReset     = "user.auth.requested_password_reset.v1"
	JobUserRequestedEmailVerification = "user.auth.requested_email_verifiacation.v1"
)

const (

	// Auth
	JobOrganizationAuthWildcard = "organization.auth.*.v1"

	JobOrganizationCreated     = "organization.auth.created_organization.v1"
	JobOrganizationTeamCreated = "organization.auth.created_team.v1"

	// Membership
	JobOrganizationMembershipAccepted = "organization.membership.accepted_user.v1"
	JobOrganizationMembershipInvited  = "organization.membership.invited_user.v1"

	// Roles
	JobOrganizationRoleWildcard = "organization.roles.*.v1"

	JobOrganizationMemRoleAssigned     = "organization.roles.assigned_membership_role.v1"
	JobOrganizationMemRoleUpdated      = "organization.roles.updated_membership_role.v1"
	JobOrganizationMemRoleRevoked      = "organization.roles.revoked_membership_role.v1"
	JobOrganizationTeamMemRoleAssigned = "organization.roles.assigned_team_membership_role.v1"
	JobOrganizationTeamMemRoleUpdated  = "organization.roles.updated_team_membership_role.v1"
	JobOrganizationTeamMemRoleRevoked  = "organization.roles.revoked_team_membership_role.v1"
)
