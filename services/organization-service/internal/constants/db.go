// Package constants
package constants

// Organization Membership Table
const (
	OrgMemStatusActive    = "active"
	OrgMemStatusLeft      = "left"
	OrgMemStatusSuspended = "suspended"
)

func IsValidaOrgMemStatus(status string) bool {
	return status == "active" || status == "suspended" || status == "left"
}
