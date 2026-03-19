// Package account contains the business logic for account
package account

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	accountv1 "github.com/rijum8906/relay/packages/pb/user_service/account/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type AccountService interface {
	// Core Functions
	MyAccount(ctx context.Context, authzMetadata *request.AuthzMetadata) (*db.Account, *errors.AppError)
	UpdateEmail(ctx context.Context, req *accountv1.UpdateEmailRequest, authzMetadata *request.AuthzMetadata, email string) *errors.AppError
	IsEmailExists(ctx context.Context, email string) (bool, *errors.AppError)
}

type accountService struct {
	repo        *accountRepository
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewAccountService(repo *accountRepository, queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) AccountService {
	return &accountService{
		repo: repo,
		q:    queries,
		utilsConfig: &utils.UtilsConfig{
			HashService:      cfg.HashService,
			JwtService:       cfg.JwtService,
			SecureJWTService: cfg.SecureJWTService,
			Storage:          cfg.Storage,
		},
		env: env,
	}
}

type AccountRepository interface {
	CreateAccount(ctx context.Context, data *authv1.SignupRequest) (*db.Account, *errors.AppError)
	GetAccount(ctx context.Context, id pgtype.UUID) (db.Account, *errors.AppError)
	GetAccountByEmail(ctx context.Context, email string) (db.Account, *errors.AppError)
	UpdatePassword(ctx context.Context, newPassword string, authzMetadata *request.AuthzMetadata) *errors.AppError
	UpdateEmail(ctx context.Context, newEmail string, authzMetadata *request.AuthzMetadata) *errors.AppError
	DeleteAccount(ctx context.Context, authzMetadata *request.AuthzMetadata) *errors.AppError
}

type accountRepository struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewAccountRepository(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) AccountRepository {
	return &accountRepository{
		q: queries,
		utilsConfig: &utils.UtilsConfig{
			HashService:      cfg.HashService,
			JwtService:       cfg.JwtService,
			SecureJWTService: cfg.SecureJWTService,
			Storage:          cfg.Storage,
		},
		env: env,
	}
}
