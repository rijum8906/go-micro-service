// Package helper
package helper

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/app"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"go.uber.org/zap"
)

var (
	instance  *ServiceHelper
	once      sync.Once
	helperErr *apperror.AppError
)

type ServiceHelper struct {
	DBPool              *pgxpool.Pool
	DBQ                 *db.Queries
	TuppleManager       coreopenfga.TuppleManager
	PermissionManager   *permissions.PermissionManager
	Logger              *zap.Logger
	OrgOpenFGAPublisher broker.Publisher
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

	publisher := broker.NewPublisher(application.BrokerCLient().GetClient())

	instance = &ServiceHelper{
		DBPool:              application.DB(),
		DBQ:                 q,
		TuppleManager:       tuppleManager,
		OrgOpenFGAPublisher: publisher,
		PermissionManager:   permissionManager,
	}
	return nil, nil
}

func GetHelper() (*ServiceHelper, *apperror.AppError) {
	if instance != nil {
		return instance, nil
	}
	once.Do(func() {
		instance, helperErr = initHelper()
	})

	if helperErr != nil {
		return nil, helperErr
	}
	if instance == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize service helper")
	}

	return instance, nil
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
			err = apperror.ErrInternal.WithMessage("panic in transaction")
			if s.Logger != nil {
				s.Logger.Error("panic in transaction",
					zap.Any("panic", p),
					zap.String("stack", string(debug.Stack())))
			}

			if rbErr := tx.Rollback(ctx); rbErr != nil {
				if s.Logger != nil {
					s.Logger.Error("rollback failed after panic", zap.Error(rbErr))
				}
			}
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				if s.Logger != nil {
					s.Logger.Warn("rollback failed", zap.Error(rbErr))
				}
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
