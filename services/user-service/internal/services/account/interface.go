package account

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/response"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

type AccountService interface {
	Signout(ctx context.Context, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) *errors.AppError
	CheckAccountExist(ctx context.Context, id pgtype.UUID) (*response.CheckAccountExistResult, *errors.AppError)
	DeleteAccount(ctx context.Context, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) *errors.AppError
	MyAccount(ctx context.Context, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*response.MyAccountResult, *errors.AppError)
	ChangeEmail(ctx context.Context, data request.ChangeEmailRequest, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*response.ChangeEmailResult, *errors.AppError)
	ChangePassword(ctx context.Context, data request.ChangePasswordRequest, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) *errors.AppError
	GenerateScopedToken(ctx context.Context, data request.GenerateScopedTokenRequest, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*response.GenerateScopedTokenResult, *errors.AppError)
}

type accountService struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewAccountService(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) AccountService {
	return &accountService{
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
