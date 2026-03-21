package mocks

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	profilev1 "github.com/rijum8906/relay/packages/pb/user_service/profile/v1"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

type MockProfileRepo struct {
	CreateProfileFunc          func(ctx context.Context, accountID pgtype.UUID, data *authv1.SignupRequest) (*db.Profile, *errors.AppError)
	GetProfilesByAccountIDFunc func(ctx context.Context, accountID pgtype.UUID) (*[]db.Profile, *errors.AppError)
	GetProfileFunc             func(ctx context.Context, profileID pgtype.UUID) (*db.Profile, *errors.AppError)
	UpdateProfileFunc          func(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateProfileRequest) (*db.Profile, *errors.AppError)
}

func (m *MockProfileRepo) CreateProfile(ctx context.Context, accountID pgtype.UUID, data *authv1.SignupRequest) (*db.Profile, *errors.AppError) {
	return m.CreateProfileFunc(ctx, accountID, data)
}

func (m *MockProfileRepo) GetProfilesByAccountID(ctx context.Context, accountID pgtype.UUID) (*[]db.Profile, *errors.AppError) {
	return m.GetProfilesByAccountIDFunc(ctx, accountID)
}

func (m *MockProfileRepo) GetProfile(ctx context.Context, profileID pgtype.UUID) (*db.Profile, *errors.AppError) {
	return m.GetProfileFunc(ctx, profileID)
}

func (m *MockProfileRepo) UpdateProfile(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateProfileRequest) (*db.Profile, *errors.AppError) {
	return m.UpdateProfileFunc(ctx, profileID, data)
}
