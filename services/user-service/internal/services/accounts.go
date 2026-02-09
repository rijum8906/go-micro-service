package services

import (
	"context"

	"github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
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
