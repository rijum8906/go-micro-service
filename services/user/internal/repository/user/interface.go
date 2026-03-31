// Package user
package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/dto"
)

type UserRepository interface {
	CreateUser(ctx context.Context, data *dto.Register) (*db.User, *apperror.AppError)
	UpdateUserPassword(ctx context.Context, id uuid.UUID, newPassHash string) *apperror.AppError
	UpdateUserEmail(ctx context.Context, id uuid.UUID, email string) *apperror.AppError
	VerifyUserEmail(ctx context.Context, id uuid.UUID) *apperror.AppError
	GetUser(ctx context.Context, id uuid.UUID) (*db.User, *apperror.AppError)
	GetUserByEmail(ctx context.Context, email string) (*db.User, *apperror.AppError)
	DeleteUser(ctx context.Context, id uuid.UUID) *apperror.AppError
}

type userRepository struct {
	q db.Querier
}

func NewAuthRepository(q db.Querier) UserRepository {
	return &userRepository{
		q: q,
	}
}
