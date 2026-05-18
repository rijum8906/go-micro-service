// Package helper
package helper

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/app"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"go.uber.org/zap"
)

var (
	instance  *ServiceHelper
	once      sync.Once
	helperErr *apperror.AppError
)

type ServiceHelper struct {
	DBPool            *pgxpool.Pool
	DBQ               *db.Queries
	TuppleManager     coreopenfga.TuppleManager
	PermissionManager *permissions.PermissionManager
	Logger            *zap.Logger
}

func initHelper() (*ServiceHelper, *apperror.AppError) {
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
	instance = &ServiceHelper{
		DBPool:            application.DB(),
		DBQ:               q,
		TuppleManager:     tuppleManager,
		PermissionManager: permissionManager,
	}
	return nil, nil
}

func GetHelper() (*ServiceHelper, *apperror.AppError) {
	once.Do(func() {
		instance, helperErr = initHelper()
	})

	if helperErr != nil {
		return nil, helperErr
	}

	return instance, nil
}

// RemoveRole removes the role from the user
// whether the role is a standard role or custom role
func (s *ServiceHelper) RemoveRole(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
	if constants.IsStandardOrgRole(targetMembership.Role) {
		s.TuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: targetMembership.Role,
				Object:   "organization:" + targetMembership.OrganizationID.String(),
			},
		})
	} else {
		s.TuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: "allowed",
				Object:   permissions.GenerateCustomRoleObject(targetMembership.OrganizationID.String(), targetMembership.Role),
			},
		})
	}
	return nil
}

func (s *ServiceHelper) AddRole(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
	if constants.IsStandardOrgRole(targetMembership.Role) {
		s.TuppleManager.Write(ctx, []client.ClientTupleKey{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: targetMembership.Role,
				Object:   "organization:" + targetMembership.OrganizationID.String(),
			},
		})
	} else {
		s.TuppleManager.Write(ctx, []client.ClientTupleKey{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: "allowed",
				Object:   permissions.GenerateCustomRoleObject(targetMembership.OrganizationID.String(), targetMembership.Role),
			},
		})
	}
	return nil
}

func (s *ServiceHelper) RunInTx(ctx context.Context, f func(q *db.Queries) *apperror.AppError) (err error) {
	tx, err := s.DBPool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to begin transaction").WithDetail("error", err.Error())
	}

	defer func() {
		if p := recover(); p != nil {
			// Log the panic with stack trace
			s.Logger.Error("panic in transaction",
				zap.Any("panic", p),
				zap.String("stack", string(debug.Stack())))

			// Rollback
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.Logger.Error("rollback failed after panic",
					zap.Error(rbErr))
			}
		} else if err != nil {
			// Normal error rollback
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.Logger.Warn("rollback failed",
					zap.Error(rbErr))
			}
		}
	}()

	q := s.DBQ.WithTx(tx)

	if appErr := f(q); appErr != nil {
		return appErr
	}

	if err = tx.Commit(ctx); err != nil {
		return apperror.ErrInternal.WithMessage("failed to commit transaction").WithDetail("error", err.Error())
	}

	return nil
}
