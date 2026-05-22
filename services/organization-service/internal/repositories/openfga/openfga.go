// Package openfga
package openfga

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	orgjobsdto "github.com/rijum8906/relay/packages/core/dto/jobs/organization"
	"github.com/rijum8906/relay/packages/core/jobs"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/app"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"go.uber.org/zap"
)

var (
	instance  *Repository
	once      sync.Once
	helperErr *apperror.AppError
)

type Repository struct {
	// Core
	DBPool              *pgxpool.Pool
	OrgOpenFGAPublisher broker.Publisher

	// Utils
	Logger            *zap.Logger
	DBQ               *db.Queries
	TuppleManager     coreopenfga.TuppleManager
	PermissionManager *permissions.PermissionManager
}

func initHelper() (*Repository, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}
	q := db.New(application.DB())

	fgaClient, appErr := coreopenfga.NewClient(application.Config().FGAAPIURL)
	if appErr != nil {
		return nil, appErr
	}
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)
	permissionManager := permissions.NewPermissionManager(fgaClient)

	publisher := broker.NewPublisher(application.BrokerCLient().GetClient())

	instance = &Repository{
		DBPool:              application.DB(),
		DBQ:                 q,
		TuppleManager:       tuppleManager,
		OrgOpenFGAPublisher: publisher,
		PermissionManager:   permissionManager,
	}
	return nil, nil
}

func GetHelper() (*Repository, *apperror.AppError) {
	once.Do(func() {
		instance, helperErr = initHelper()
	})

	if helperErr != nil {
		return nil, helperErr
	}

	return instance, nil
}

// PublishRevokeOrgMemRoleJob removes the role from the user
// whether the role is a standard role or custom role
func (s *Repository) PublishRevokeOrgMemRoleJob(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
	appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobOrganizationMemRoleRevoked, s.buildOrgMemRoleDTO(targetMembership))
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s *Repository) PublishAssignOrgMemRoleJob(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
	appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobOrganizationMemRoleAssigned, s.buildOrgMemRoleDTO(targetMembership))
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s Repository) PublishUpdateOrgMemRoleJob(ctx context.Context, oldMembership, newMembership *db.OrganizationMembership) *apperror.AppError {
	appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobOrganizationMemRoleUpdated, orgjobsdto.UpdateOrgRoleDTO{
		OrgRoleDTO: s.buildOrgMemRoleDTO(oldMembership),
		NewRole:    newMembership.Role,
	})
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s *Repository) PublishAssignTeamMemRoleJob(ctx context.Context, targetMembership *db.OrganizationTeamMembership) *apperror.AppError {
	appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobOrganizationMemRoleAssigned, s.buildOrgTeamRoleDTO(targetMembership))
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s *Repository) PublishRevokeTeamMemRoleJob(ctx context.Context, targetMembership *db.OrganizationTeamMembership) *apperror.AppError {
	appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobOrganizationMemRoleRevoked, s.buildOrgTeamRoleDTO(targetMembership))
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s *Repository) PublishUpdateTeamMemRoleJob(ctx context.Context, oldMembership, newMembership *db.OrganizationTeamMembership) *apperror.AppError {
	appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobOrganizationMemRoleUpdated, orgjobsdto.UpdateOrgRoleDTO{
		OrgRoleDTO: s.buildOrgTeamRoleDTO(oldMembership),
		NewRole:    newMembership.Role,
	})
	if appErr != nil {
		return appErr
	}
	return nil
}
