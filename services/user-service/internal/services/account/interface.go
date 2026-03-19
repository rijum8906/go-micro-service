package account

import (
	"context"

	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type AccountService interface {
	GenerateScopedToken(ctx context.Context, req *user_servicev1.GenerateScopedTokenRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.GenerateScopedTokenResponse, *errors.AppError)
	ChangePassword(ctx context.Context, req *user_servicev1.ChangePasswordRequest, authzMetadata request.AuthzMetadata) *errors.AppError
	DeleteAccount(ctx context.Context, req *user_servicev1.DeleteAccountRequest, authzMetadata request.AuthzMetadata) *errors.AppError
	MyAccount(ctx context.Context, req *user_servicev1.GetMyAccountRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.GetMyAccountResponse, *errors.AppError)
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
