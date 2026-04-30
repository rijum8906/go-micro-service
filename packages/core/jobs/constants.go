// Package jobs
package jobs

import (
	"fmt"
	"strings"
)

// NOTE: naming convension of jobs - [domain].[subdomain].[event].[version]
// eg. user.auth.requested_password_reset.v1, user.auth.created_user.v1, user.auth.updated_user.v1
// 	user.security.email_changed.v1, user.security.password_changed.v1
// 	organization.auth.created_organization.v1, organization.membership.invited_user.v1
// 	organization.billing.subscroption_started.v1
//
// NOTE: domain could be the name of the service
// all domain must be declared here
// subdomain could be anything and will be decided by the service
// event must be in past form and the name should be clear

type Domain string

const (
	DomainUser         Domain = "user"
	DomainTask         Domain = "task"
	DomainNotification Domain = "notification"
	DomainOrganization Domain = "organization"
)

var Domains = []Domain{
	DomainUser,
	DomainTask,
	DomainNotification,
	DomainOrganization,
}

// GetDomainWildcard returns wildcard for entire domain with version
func GetDomainWildcard(job string) string {
	parts := strings.Split(job, ".")
	return fmt.Sprintf("%s.*.*.%s", parts[0], parts[3])
}

func GetDomainWildcardWithoutVersion(job string) string {
	parts := strings.Split(job, ".")
	return fmt.Sprintf("%s.*.*.*", parts[0])
}

// GetSubdomainWildcard returns wildcard for entire job with version
func GetSubdomainWildcard(job string) string {
	parts := strings.Split(job, ".")
	return fmt.Sprintf("%s.%s.*.%s", parts[0], parts[1], parts[3])
}

func GetSubdomainWildcardWithoutVersion(job string) string {
	parts := strings.Split(job, ".")
	return fmt.Sprintf("%s.%s.*.*", parts[0], parts[1])
}
