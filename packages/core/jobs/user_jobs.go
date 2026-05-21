package jobs

const (
	JobUserAuthWildcard               = "user.auth.*.v1"
	JobUserRequestedPasswordReset     = "user.auth.requested_password_reset.v1"
	JobUserRequestedEmailVerification = "user.auth.requested_email_verifiacation.v1"
)

const (
	JobOrganizationAuthWildcard      = "organization.auth.*.v1"
	JobOrganizationCreated           = "organization.auth.created_organization.v1"
	JobOrganizationTeamCreated       = "organization.auth.created_team.v1"
	JobOrganizationMembershipInvited = "organization.membership.invited_user.v1"
)
