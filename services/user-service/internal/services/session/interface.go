// Package session
package session

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/hash"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/token"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app"
	"github.com/rijum8906/relay/services/user/app/config"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/services/helper"
	"go.uber.org/zap"
)

type SessionService struct {
	// Core
	DBPool              *pgxpool.Pool
	DBQ                 *db.Queries
	UserClient          userv1.UserServiceClient
	OrgOpenFGAPublisher broker.Publisher
	Helper              *helper.ServiceHelper

	// Utils
	TuppleManager coreopenfga.TuppleManager
	TokenManager  token.TokenManager
	Permission    *permissions.PermissionManager
	HashService   *hash.HashService
	Logger        *zap.Logger

	// Config
	Config *config.Env
}

func New() (sessionv1.SessionServiceServer, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}

	q := db.New(application.DB())

	fgaClient, appErr := coreopenfga.NewClient(application.Config().FGAAPIURL)
	if appErr != nil {
		return nil, appErr
	}
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)
	permissionManager := permissions.NewPermissionManager(fgaClient)

	publisher := broker.NewPublisher(application.BrokerCLient().GetClient())

	helper, appErr := helper.GetHelper()
	if appErr != nil {
		return nil, appErr
	}

	return &SessionService{
		DBPool:              application.DB(),
		DBQ:                 q,
		TuppleManager:       tuppleManager,
		TokenManager:        application.TokenManager(),
		Permission:          permissionManager,
		OrgOpenFGAPublisher: publisher,
		HashService: hash.NewHashService(hash.Config{
			BcryptCost: 10,
		}),
		Logger: application.Logger(),
		Config: application.Config(),
		Helper: helper,
	}, nil
}
