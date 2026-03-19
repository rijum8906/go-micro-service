package account

import (
	"context"
	"database/sql"
	errorsstdlib "errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

func (r *accountRepository) IsEmailExists(ctx context.Context, email string) (bool, *errors.AppError) {
	_, err := r.q.GetAccountByEmail(ctx, email)
	if err != nil {
		if errorsstdlib.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.ErrInternal.WithInternal(err)
	}
	return true, nil
}

func (r *accountRepository) CreateAccount(ctx context.Context, data *authv1.SignupRequest) (*db.Account, *errors.AppError) {
	passwordHash, err := r.utilsConfig.HashService.HashPassword(data.Password.Value)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	account, err := r.q.CreateAccount(ctx, db.CreateAccountParams{
		Email:        data.Email.Value,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &account, nil
}

func (r *accountRepository) GetAccount(ctx context.Context, id pgtype.UUID) (db.Account, *errors.AppError) {
	account, err := r.q.GetAccount(ctx, id)
	if err != nil {
		return db.Account{}, errors.ErrInternal.WithInternal(err)
	}
	return account, nil
}

func (r *accountRepository) GetAccountByEmail(ctx context.Context, email string) (db.Account, *errors.AppError) {
	account, err := r.q.GetAccountByEmail(ctx, email)
	if err != nil {
		return db.Account{}, errors.ErrInternal.WithInternal(err)
	}
	return account, nil
}

func (r *accountRepository) UpdatePassword(ctx context.Context, newPassword string, authzMetadata *request.AuthzMetadata) *errors.AppError {
	passwordHash, err := r.utilsConfig.HashService.HashPassword(newPassword)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}

	_, err = r.q.UpdateAccount(ctx, db.UpdateAccountParams{
		PasswordHash: pgtype.Text{
			Valid:  true,
			String: passwordHash,
		},
	})
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}

	return nil
}

func (r *accountRepository) UpdateEmail(ctx context.Context, newEmail string, authzMetadata *request.AuthzMetadata) *errors.AppError {
	_, err := r.q.UpdateAccount(ctx, db.UpdateAccountParams{
		Email: pgtype.Text{Valid: true, String: newEmail},
	})
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (r *accountRepository) DeleteAccount(ctx context.Context, authzMetadata *request.AuthzMetadata) *errors.AppError {
	err := r.q.DeleteAccount(ctx, authzMetadata.UserID)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}
