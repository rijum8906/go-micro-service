package app

import (
	"context"
	"net"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ApplicationInfra struct {
	cache        *redis.Client
	database     *pgxpool.Pool
	brokerClient broker.Client
	fgaClient    *coreopenfga.Client
}

type ApplicationUtils struct {
	logger *zap.Logger
}

type ApplicationServices struct {
	OrganizationService organizationv1.OrganizationServiceServer
}

type GrpcClients struct {
	UserClient userv1.UserServiceClient
}

type Application struct {
	config   *config.Env
	infra    *ApplicationInfra
	utils    *ApplicationUtils
	services *ApplicationServices
	clients  *GrpcClients
	listener net.Listener
	server   *grpc.Server
}

var (
	instance *Application
	once     sync.Once
	mu       sync.RWMutex
	initErr  *apperror.AppError
)

// GetInstance safely fetches the singleton instance.
// Thread-safe via double-checked locking idiom combined with sync.Once.
func GetInstance() (*Application, *apperror.AppError) {
	// Fast path: Check if already initialized using an atomic read lock
	mu.RLock()
	inst := instance
	err := initErr
	mu.RUnlock()

	if inst != nil {
		return inst, nil
	}
	if err != nil {
		return nil, InternalError(err)
	}

	// Slow path: Initialize under a total write lock to prevent ResetInstance data races
	mu.Lock()
	defer mu.Unlock()

	once.Do(func() {
		ctx := context.Background()
		instance, initErr = newApplication(ctx)
	})

	if initErr != nil {
		return nil, InternalError(initErr)
	}

	return instance, nil
}

// NewApplication is now unexported. Global application creation should exclusively
// flow through GetInstance to guarantee singleton sanctity.
func newApplication(ctx context.Context) (*Application, *apperror.AppError) {
	app := &Application{
		infra:    &ApplicationInfra{},
		utils:    &ApplicationUtils{},
		services: &ApplicationServices{},
		clients:  &GrpcClients{},
	}

	var appErr *apperror.AppError

	app.config, appErr = config.LoadEnv()
	if appErr != nil {
		return nil, appErr
	}

	// Initialize the utils forst (logger)
	if appErr = app.initUtils(); appErr != nil {
		return nil, appErr
	}

	// Set the error config
	apperror.SetConfig(apperror.Config{
		AppEnv: app.config.AppEnv,
		Debug:  true,
		Logger: app.utils.logger,
	})

	if appErr = app.initInfra(ctx); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initGRPCClients(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initServices(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initGRPCServer(); appErr != nil {
		return nil, appErr
	}

	return app, nil
}

// ResetInstance safely resets the singleton instance under a full Mutex lock.
// This prevents concurrent data races on the `once` synchronization primitive.
func ResetInstance() {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		instance.cleanupInternal()
		instance = nil
	}
	initErr = nil
	once = sync.Once{}
}

// Cleanup is the thread-safe public hook for graceful shutdowns.
func (a *Application) Cleanup() {
	mu.Lock()
	defer mu.Unlock()
	a.cleanupInternal()
}

func (a *Application) cleanupInternal() {
	if a.server != nil {
		a.server.GracefulStop()
	}

	if a.listener != nil {
		a.listener.Close()
	}

	if a.infra != nil {
		if a.infra.database != nil {
			a.infra.database.Close()
		}
		if a.infra.cache != nil {
			a.infra.cache.Close()
		}
		if a.infra.brokerClient != nil {
			a.infra.brokerClient.Close()
		}
		if a.infra.fgaClient != nil && a.infra.fgaClient.Client != nil {
			a.infra.fgaClient.Client.GetConfig().HTTPClient.CloseIdleConnections()
		}
	}

	if a.utils != nil && a.utils.logger != nil {
		_ = a.utils.logger.Sync()
	}
}

// Thread-safe Exported Getters to access unexported infrastructure segments smoothly

// IsInitialized checks if the application is initialized safely using read locks.
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return instance != nil && initErr == nil
}

func InternalError(err error) *apperror.AppError {
	if appErr, ok := err.(*apperror.AppError); ok {
		return appErr
	}
	return apperror.ErrInternal.WithDetail("error", err.Error())
}
