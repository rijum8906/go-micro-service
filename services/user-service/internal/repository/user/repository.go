package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
)

func (r *userRepository) CreateUser(ctx context.Context, data *authv1.RegisterRequest) (*db.User, *apperror.AppError) {
	user, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email: data.Email,
		PasswordHash: pgtype.Text{
			Valid:  true,
			String: data.Password,
		},
		TwoFactorEnabledAt: pgtype.Timestamptz{},
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to create user").WithDetail("error", err.Error())
	}
	return &user, nil
}

func (r *userRepository) GetUser(ctx context.Context, id uuid.UUID) (*db.User, *apperror.AppError) {
	user, err := r.q.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("User not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to get user").WithDetail("error", err.Error())
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, *apperror.AppError) {
	user, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("User not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to get user").WithDetail("error", err.Error())
	}
	return &user, nil
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, id uuid.UUID, newPassHash string) *apperror.AppError {
	_, err := r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID: id,
		PasswordHash: pgtype.Text{
			Valid:  true,
			String: newPassHash,
		},
	})
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to update user").WithDetail("error", err.Error())
	}
	return nil
}

func (r *userRepository) UpdateUserEmail(ctx context.Context, id uuid.UUID, email string) *apperror.AppError {
	_, err := r.q.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    id,
		Email: email,
	})
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to update user").WithDetail("error", err.Error())
	}
	return nil
}

func (r *userRepository) VerifyUserEmail(ctx context.Context, id uuid.UUID) *apperror.AppError {
	_, appErr := r.GetUser(ctx, id)
	if appErr != nil {
		return appErr
	}

	_, err := r.q.UpdateUserIsEmailVerified(ctx, db.UpdateUserIsEmailVerifiedParams{
		ID:              id,
		IsEmailVerified: true,
		EmailVerifiedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to update user").WithDetail("error", err.Error())
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) *apperror.AppError {
	err := r.q.DeleteUser(ctx, id)
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to delete user").WithDetail("error", err.Error())
	}
	return nil
}

func (r *userRepository) CheckExists(ctx context.Context, id uuid.UUID) (bool, *apperror.AppError) {
	exists, err := r.q.CheckUserExists(ctx, id)
	if err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return exists, nil
}

func (r *userRepository) CheckEmailExists(ctx context.Context, email string) (bool, *apperror.AppError) {
	exists, err := r.q.CheckUserEmailExists(ctx, email)
	if err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return exists, nil
}
