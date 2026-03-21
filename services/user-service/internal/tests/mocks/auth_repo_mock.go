// Package mocks contains mock implementations for the repos
package mocks

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

type MockAuthRepo struct {
	CreateSessionFunc            func(ctx context.Context, refreshToken string, metadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError)
	GetSessionByRefreshTokenFunc func(ctx context.Context, refreshToken string, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError)
	RevokeSessionFunc            func(ctx context.Context, id pgtype.UUID, authzMetadata request.AuthzMetadata) *errors.AppError
	RevokeAllSessionsFunc        func(ctx context.Context, authzMetadata request.AuthzMetadata) *errors.AppError
}

func (m *MockAuthRepo) CreateSession(ctx context.Context, refreshToken string, metadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError) {
	return m.CreateSessionFunc(ctx, refreshToken, metadata, authzMetadata)
}

func (m *MockAuthRepo) GetSessionByRefreshToken(ctx context.Context, refreshToken string, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError) {
	return m.GetSessionByRefreshTokenFunc(ctx, refreshToken, authzMetadata)
}

func (m *MockAuthRepo) RevokeSession(ctx context.Context, id pgtype.UUID, authzMetadata request.AuthzMetadata) *errors.AppError {
	return m.RevokeSessionFunc(ctx, id, authzMetadata)
}

func (m *MockAuthRepo) RevokeAllSessions(ctx context.Context, authzMetadata request.AuthzMetadata) *errors.AppError {
	return m.RevokeAllSessionsFunc(ctx, authzMetadata)
}
