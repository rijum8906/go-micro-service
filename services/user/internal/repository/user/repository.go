package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/dto"
)

func (r *userRepository) CreateUser(ctx context.Context, data *dto.Register) (*db.User, *apperror.AppError) {
	user, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email: data.Email,
		PasswordHash: pgtype.Text{
			Valid:  true,
			String: data.Password,
		},
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
	_, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
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
	_, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:    id,
		Email: email,
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
