package account

import (
	"github.com/rijum8906/relay/packages/common/env"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type AccountService interface {
	// Core Functions
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

type AccountRepository interface {
	CreateAccount()
	GetAccount()
	GetAccountByEmail()
	UpdatePassword()
	UpdateEmail()
	DeleteAccount()
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
