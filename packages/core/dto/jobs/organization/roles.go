// Package orgjobsdto
package orgjobsdto

type OrgRoleDTO struct {
	User         string
	Role         string
	Organization string
}

type UpdateOrgRoleDTO struct {
	OrgRoleDTO
	NewRole string
}

type TeamRoleDTO struct {
	User string
	Role string
	Team string
}

type UpdateTeamRoleDTO struct {
	TeamRoleDTO
	NewRole string
}
