// Package constants
package constants

import "slices"

// Organization Membership Table
const (
	// Statues
	OrgMemStatusActive    = "active"
	OrgMemStatusLeft      = "left"
	OrgMemStatusDeleted   = "left"
	OrgMemStatusRemoved   = "removed"
	OrgMemStatusSuspended = "suspended"
	OrgMemStatusBanned    = "banned"
)

var OrgMemStatuses = []string{OrgMemStatusActive, OrgMemStatusLeft, OrgMemStatusDeleted, OrgMemStatusRemoved, OrgMemStatusSuspended, OrgMemStatusBanned}

func IsValidaOrgMemStatus(status string) bool {
	return slices.Contains(OrgMemStatuses, status)
}

// Organization Table
const (
	// Statues
	OrgStatusActive   = "active"
	OrgStatusArchived = "archived"
	OrgStatusDeleted  = "deleted"

	// Roles
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"
)

var (
	OrgStatuses = []string{OrgStatusActive, OrgStatusArchived, OrgStatusDeleted}
	OrgRoles    = []string{OrgRoleOwner, OrgRoleAdmin, OrgRoleMember}
)

func IsValidaOrgStatus(status string) bool {
	return slices.Contains(OrgStatuses, status)
}

func IsStandardOrgRole(role string) bool {
	return slices.Contains(OrgRoles, role)
}
