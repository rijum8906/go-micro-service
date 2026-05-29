// Package helper
package helper

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/app"
	"github.com/rijum8906/relay/services/user/internal/db"
	"go.uber.org/zap"
)

var (
	instance  *ServiceHelper
	once      sync.Once
	helperErr *apperror.AppError
)

type ServiceHelper struct {
	// Core
	DBPool *pgxpool.Pool
	DBQ    *db.Queries

	// Utils
	Logger *zap.Logger
}

func initHelper() (*ServiceHelper, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}
	q := db.New(application.DB())

	instance = &ServiceHelper{
		DBPool: application.DB(),
		DBQ:    q,
		Logger: application.Logger(),
	}
	return instance, nil
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

func GetHelperForTest(pool *pgxpool.Pool, q *db.Queries, logger *zap.Logger) *ServiceHelper {
	return &ServiceHelper{
		DBPool: pool,
		DBQ:    q,
		Logger: logger,
	}
}

func (s *ServiceHelper) RunInTx(ctx context.Context, f func(q *db.Queries) *apperror.AppError) *apperror.AppError {
	tx, err := s.DBPool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to begin transaction").WithDetail("error", err.Error())
	}

	defer func() {
		if p := recover(); p != nil {
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

func (s *ServiceHelper) RunInTxWithManualRollback(ctx context.Context, f func(q *db.Queries) *apperror.AppError) (pgx.Tx, *apperror.AppError) {
	tx, err := s.DBPool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to begin transaction").WithDetail("error", err.Error())
	}

	q := s.DBQ.WithTx(tx)
	if appErr := f(q); appErr != nil {
		return nil, appErr
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to commit transaction").WithDetail("error", err.Error())
	}

	return tx, nil
}
