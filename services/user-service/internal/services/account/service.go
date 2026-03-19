package account

import (
	"context"

	"github.com/rijum8906/relay/packages/common/errors"
	accountv1 "github.com/rijum8906/relay/packages/pb/user_service/account/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

func (s *accountService) MyAccount(ctx context.Context, authzMetadata *request.AuthzMetadata) (*db.Account, *errors.AppError) {
	account, err := s.q.GetAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return &account, nil
}

func (s *accountService) UpdateEmail(ctx context.Context, req *accountv1.UpdateEmailRequest, authzMetadata *request.AuthzMetadata, email string) *errors.AppError {
	appErr := s.repo.UpdateEmail(ctx, req.NewEmail.Value, authzMetadata)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s *accountService) IsEmailExists(ctx context.Context, email string) (bool, *errors.AppError) {
	_, appErr := s.repo.GetAccountByEmail(ctx, email)
	if appErr != nil {
		return false, appErr
	}
	return true, nil
}
