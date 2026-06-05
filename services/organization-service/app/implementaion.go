// Package app
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/corelogger"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/organization"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// NOTE: do not use the logger here

func (a *Application) initInfra(ctx context.Context) *apperror.AppError {
	// PostgreSQL
	database, appErr := initDB(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.database = database

	// Redis
	cache, appErr := initCache(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.cache = cache

	// NATS
	nats, appErr := initNATSClient(a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.brokerClient = nats

	// OpenFGA
	fgaClient, appErr := initFgaClient(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.fgaClient = fgaClient

	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	logger, appErr := corelogger.InitLogger(corelogger.LoggerConfig{
		AppEnv:       a.config.AppEnv,
		LogLevel:     a.config.LogLevel,
		LogFile:      a.config.LogFile,
		EnableJSON:   a.config.EnableJSON,
		EnableCaller: a.config.EnableCaller,
		EnableStack:  a.config.EnableStack,
	})
	if appErr != nil {
		return appErr
	}
	a.utils.logger = logger

	return nil
}

func (a *Application) initGRPCClients() *apperror.AppError {
	conn, err := grpc.NewClient(a.config.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return apperror.ErrThirdParty.WithMessage("failed to connect to user service").WithDetail("error", err.Error())
	}

	// connect to clients
	a.clients = &GrpcClients{
		UserClient: userv1.NewUserServiceClient(conn),
	}
	return nil
}

func (a *Application) initGRPCServer() *apperror.AppError {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.Port))
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to listen for gRPC server").WithDetail("error", err.Error())
	}
	a.listener = listener

	// Create and Register grpc server
	server := grpc.NewServer()
	organizationv1.RegisterOrganizationServiceServer(server, a.services.OrganizationService)
	a.server = server

	if a.config.AppEnv == "development" {
		reflection.Register(server)
	}

	return nil
}

func (a *Application) initServices() *apperror.AppError {
	q := db.New(a.infra.database)

	a.services.OrganizationService = organization.New(q, a.clients.UserClient, a.infra.fgaClient)

	return nil
}

func (a *Application) Run() *apperror.AppError {
	err := a.server.Serve(a.listener)
	if err != nil {
		if err == grpc.ErrServerStopped {
			return nil
		}

		return apperror.ErrInternal.WithMessage("failed to start gRPC server").WithDetail("error", err.Error())
	}

	return nil
}
