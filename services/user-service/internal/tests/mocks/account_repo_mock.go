package mocks

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

type MockAccountRepo struct {
	IsEmailExistsFunc     func(ctx context.Context, email string) (bool, *errors.AppError)
	CreateAccountFunc     func(ctx context.Context, data *authv1.SignupRequest) (*db.Account, *errors.AppError)
	GetAccountFunc        func(ctx context.Context, id pgtype.UUID) (*db.Account, *errors.AppError)
	GetAccountByEmailFunc func(ctx context.Context, email string) (*db.Account, *errors.AppError)
	UpdateEmailFunc       func(ctx context.Context, newEmail string, authzMetadata *request.AuthzMetadata) *errors.AppError
	UpdatePasswordFunc    func(ctx context.Context, newPassword string, authzMetadata *request.AuthzMetadata) *errors.AppError
	DeleteAccountFunc     func(ctx context.Context, authzMetadata *request.AuthzMetadata) *errors.AppError
}

func (m *MockAccountRepo) IsEmailExists(ctx context.Context, email string) (bool, *errors.AppError) {
	return m.IsEmailExistsFunc(ctx, email)
}

func (m *MockAccountRepo) CreateAccount(ctx context.Context, data *authv1.SignupRequest) (*db.Account, *errors.AppError) {
	return m.CreateAccountFunc(ctx, data)
}

func (m *MockAccountRepo) GetAccount(ctx context.Context, id pgtype.UUID) (*db.Account, *errors.AppError) {
	return m.GetAccountFunc(ctx, id)
}

func (m *MockAccountRepo) GetAccountByEmail(ctx context.Context, email string) (*db.Account, *errors.AppError) {
	return m.GetAccountByEmailFunc(ctx, email)
}

func (m *MockAccountRepo) UpdateEmail(ctx context.Context, newEmail string, authzMetadata *request.AuthzMetadata) *errors.AppError {
	return m.UpdateEmailFunc(ctx, newEmail, authzMetadata)
}

func (m *MockAccountRepo) UpdatePassword(ctx context.Context, newPassword string, authzMetadata *request.AuthzMetadata) *errors.AppError {
	return m.UpdatePasswordFunc(ctx, newPassword, authzMetadata)
}

func (m *MockAccountRepo) DeleteAccount(ctx context.Context, authzMetadata *request.AuthzMetadata) *errors.AppError {
	return m.DeleteAccountFunc(ctx, authzMetadata)
}
