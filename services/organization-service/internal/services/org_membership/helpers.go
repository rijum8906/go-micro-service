package orgmembership

import (
	"context"
	"runtime/debug"

	"github.com/jackc/pgx/v5"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"go.uber.org/zap"
)

// removeRole removes the role from the user
// whether the role is a standard role or custom role
func (s *OrgMembershipService) removeRole(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
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

func (s *OrgMembershipService) addRole(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
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

func (s *OrgMembershipService) runInTx(ctx context.Context, f func(q *db.Queries) *apperror.AppError) (err error) {
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
