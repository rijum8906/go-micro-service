package services

import (
	"context"
	"database/sql"
	stdErrors "errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

func (s *accountService) DeleteAccount(
	ctx context.Context,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
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
) (*dto.CheckAccountExistResult, *errors.AppError) {
	_, err := s.q.GetAccount(ctx, id)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return &dto.CheckAccountExistResult{Exist: false}, nil
		}
		return nil, errors.ErrDBError.WithInternal(err)
	}

	return &dto.CheckAccountExistResult{Exist: true}, nil
}

func (s *accountService) ChangePassword(
	ctx context.Context,
	data dto.ChangePasswordRequest,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) *errors.AppError {
	claims, appErr := s.utilsConfig.SecureJWTService.ValidateToken(ctx, data.Token)
	if appErr != nil {
		return appErr
	}

	if claims.Subject != authzMetadata.UserID.String() {
		return errors.NewAppError(http.StatusForbidden, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	if claims.Scope != jwt.ActionChangePassword {
		return errors.NewAppError(http.StatusForbidden, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	hashedPass, err := s.utilsConfig.HashService.HashPassword(data.NewPassword)
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

func (s *accountService) ChangeEmail(
	ctx context.Context,
	data dto.ChangeEmailRequest,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) (*dto.ChangeEmailResult, *errors.AppError) {
	claims, appErr := s.utilsConfig.SecureJWTService.ValidateToken(ctx, data.Token)
	if appErr != nil {
		return nil, appErr
	}

	if claims.Subject != authzMetadata.UserID.String() {
		return nil, errors.NewAppError(http.StatusForbidden, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	if claims.Scope != jwt.ActionChangeEmail {
		return nil, errors.NewAppError(http.StatusForbidden, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	_, err := s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID: authzMetadata.UserID,
		Email: pgtype.Text{
			String: data.NewEmail,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &dto.ChangeEmailResult{
		Email: data.NewEmail,
	}, nil
}

func (s *accountService) MyAccount(
	ctx context.Context,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) (*dto.MyAccountResult, *errors.AppError) {
	account, err := s.q.GetAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	profiles, err := s.q.GetProfilesByAccountID(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	accountSecuriry, err := s.q.GetAccountSecurityByAccountID(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	oAuths, err := s.q.GetOAuthsByAccountID(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	return &dto.MyAccountResult{
		Account:         &account,
		Profiles:        &profiles,
		AccountSecurity: &accountSecuriry,
		OAuths:          &oAuths,
	}, nil
}

var ActionTokens = []string{
	jwt.ActionChangeEmail,
	jwt.ActionChangePassword,
}

func (s *accountService) GenerateScopedToken(
	ctx context.Context,
	data dto.GenerateScopedTokenRequest,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) (*dto.GenerateScopedTokenResult, *errors.AppError) {
	isAvailable := false
	for _, action := range ActionTokens {
		if action == data.Scope {
			isAvailable = true
		}
	}
	if !isAvailable {
		return nil, errors.NewAppError(http.StatusForbidden, "forbidden", []errors.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	if data.Authorization.Type != dto.PassAuthzType {
		return nil, errors.NewAppError(http.StatusForbidden, "forbidden", []errors.Error{
			{Field: "auth", Message: "Other authorization types are not supported yet."},
		})
	}

	account, err := s.q.GetAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	err = s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, data.Authorization.Value)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	token, appErr := s.utilsConfig.SecureJWTService.IssueToken(ctx, jwt.ScopedActionClaims{
		Subject: authzMetadata.UserID.String(),
		Scope:   data.Scope,
	})
	if appErr != nil {
		return nil, appErr
	}
	return &dto.GenerateScopedTokenResult{
		Token: token,
	}, nil
}

func (s *accountService) Signout(ctx context.Context, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) *errors.AppError {
	redisKey := utils.GenerateRedisLoginKey(authzMetadata.UserID.String(), reqMetadata.DeviceID)
	err := s.utilsConfig.JwtService.RevokeSession(ctx, redisKey)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}
