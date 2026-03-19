// Package account contains services for the account service.
package account

import (
	"context"
	"database/sql"
	stdErrors "errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	"github.com/rijum8906/relay/packages/common/jwt"
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/response"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *accountService) DeleteAccount(
	ctx context.Context,
	req *user_servicev1.DeleteAccountRequest,
	authzMetadata request.AuthzMetadata,
) *errors.AppError {
	err := s.q.DeleteAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (s *accountService) CheckAccountExist(
	ctx context.Context,
	id pgtype.UUID,
) (*response.CheckAccountExistResult, *errors.AppError) {
	_, err := s.q.GetAccount(ctx, id)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return &response.CheckAccountExistResult{Exist: false}, nil
		}
		return nil, errors.ErrDBError.WithInternal(err)
	}

	return &response.CheckAccountExistResult{Exist: true}, nil
}

func (s *accountService) ChangePassword(
	ctx context.Context,
	req *user_servicev1.ChangePasswordRequest,
	authzMetadata request.AuthzMetadata,
) *errors.AppError {
	claims, appErr := s.utilsConfig.SecureJWTService.ValidateToken(ctx, req.Token.Value)
	if appErr != nil {
		return appErr
	}

	if claims.Subject != authzMetadata.UserID.String() {
		return errors.NewAppError(errors.ErrForbidden.Code, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	if claims.Scope != jwt.ActionChangePassword {
		return errors.NewAppError(errors.ErrForbidden.Code, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	hashedPass, err := s.utilsConfig.HashService.HashPassword(req.NewPassword.Value)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}

	_, err = s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID: authzMetadata.UserID,
		PasswordHash: pgtype.Text{
			String: hashedPass,
			Valid:  true,
		},
	})
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}

	return nil
}

func (s *accountService) MyAccount(
	ctx context.Context,
	req *user_servicev1.GetMyAccountRequest,
	authzMetadata request.AuthzMetadata,
) (*user_servicev1.GetMyAccountResponse, *errors.AppError) {
	account, err := s.q.GetAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	accountSecuriry, err := s.q.GetAccountSecurityByAccountID(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	return &user_servicev1.GetMyAccountResponse{
		Id:    utils.NewID(account.ID.String()),
		Email: utils.NewEmail(account.Email),
		Security: &user_servicev1.GetMyAccountSecurityResponse{
			IsEmailVerified:    accountSecuriry.IsEmailVerified,
			EmailVerifiedAt:    utils.NewTimestamp(accountSecuriry.EmailVerifiedAt.Time),
			TwoFactorEnabled:   accountSecuriry.TwoFactorEnabled,
			TwoFactorEnabledAt: utils.NewTimestamp(accountSecuriry.TwoFactorEnabledAt.Time),
		},
		CreatedAt: utils.NewTimestamp(account.CreatedAt.Time),
		UpdatedAt: utils.NewTimestamp(account.UpdatedAt.Time),
	}, nil
}

func (s *accountService) GenerateScopedToken(
	ctx context.Context,
	req *user_servicev1.GenerateScopedTokenRequest,
	authzMetadata request.AuthzMetadata,
) (*user_servicev1.GenerateScopedTokenResponse, *errors.AppError) {
	isAvailable := false
	for _, action := range user_servicev1.ScopedAction_name {
		if action == req.Scope.String() {
			isAvailable = true
		}
	}
	if !isAvailable {
		return nil, errors.NewAppError(errors.ErrForbidden.Code, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	if req.AuthType.String() != user_servicev1.AuthType_AUTH_TYPE_PASSWORD.String() {
		return nil, errors.NewAppError(errors.ErrForbidden.Code, "forbidden", []errors.Error{
			{Field: "auth", Message: "Other authorization types are not supported yet."},
		})
	}

	account, err := s.q.GetAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	err = s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, req.Auth.Value)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	token, appErr := s.utilsConfig.SecureJWTService.IssueToken(ctx, jwt.ScopedActionClaims{
		Subject: authzMetadata.UserID.String(),
		Scope:   req.Scope.String(),
	})
	if appErr != nil {
		return nil, appErr
	}

	return &user_servicev1.GenerateScopedTokenResponse{
		Token: &commonv1.Token{
			Value: token,
			ExpiresAt: &timestamppb.Timestamp{
				Seconds: int64(s.env.ScopedJwtExpiration.Seconds()),
			},
		},
	}, nil
}
