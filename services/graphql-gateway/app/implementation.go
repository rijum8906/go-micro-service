package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *Application) initCache(ctx context.Context) *apperror.AppError {
	cache, appErr := cache.Connect(ctx, cache.Config{
		Host:        a.config.RedisHost,
		Port:        a.config.RedisPort,
		DB:          0,
		Password:    a.config.RedisPass,
		RetryCounts: 5,
	})

	if appErr != nil {
		return appErr
	}

	a.infra.cache = cache
	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(a.config.JWTSecret),
		ScopedSecret:   []byte(a.config.ScopedSecret),
		SessionTTL:     a.config.SessionTTL,
		ScopedTokenTTL: a.config.ScopedTokenTTL,
	}, a.infra.cache)
	a.utils.token = tokenManager

	return nil
}

func (a *Application) initGRPCClients() *apperror.AppError {
	conn, err := grpc.NewClient(a.config.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return apperror.ErrThirdParty.WithMessage("failed to connect to user service").WithDetail("error", err.Error())
	}

	// Initialize gRPC clients
	authClient := authv1.NewAuthServiceClient(conn)
	a.clients = &GrpcClients{
		AuthConn:   conn,
		AuthClient: authClient,
	}

	return nil
}

func (a *Application) Run() *apperror.AppError {
	err := a.server.Serve(a.listener)
	if err != nil && err != http.ErrServerClosed {
		return apperror.ErrInternal.WithMessage("failed to start HTTP server").WithDetail("error", err.Error())
	}

	return nil
}

func (a *Application) Shutdown(ctx context.Context) {
	if a.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
	}

	if a.clients != nil && a.clients.AuthConn != nil {
		_ = a.clients.AuthConn.Close()
	}

	if a.infra != nil && a.infra.cache != nil {
		_ = a.infra.cache.Close()
	}
}

func (a *Application) port() string {
	return strconv.Itoa(a.config.Port)
}

func (a *Application) Addr() string {
	return fmt.Sprintf(":%s", a.port())
}

func (a *Application) UserServiceAddr() string {
	return fmt.Sprintf("%s", a.config.UserServiceAddr)
}
