// Package app
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/app/config"
	taskdb "github.com/rijum8906/relay/services/task-service/internal/db"
	handler "github.com/rijum8906/relay/services/task-service/internal/handlers/grpc"
	projectRepo "github.com/rijum8906/relay/services/task-service/internal/repository/project"
	projectmembershiprepo "github.com/rijum8906/relay/services/task-service/internal/repository/project_membership"
	taskrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task"
	taskassignmentrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task_assignment"
	taskcommentrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task_comment"
	projectservice "github.com/rijum8906/relay/services/task-service/internal/services/project"
	projectmembershipservice "github.com/rijum8906/relay/services/task-service/internal/services/project_membership"
	taskservice "github.com/rijum8906/relay/services/task-service/internal/services/task"
	taskassigmentservice "github.com/rijum8906/relay/services/task-service/internal/services/task_assigment"
	taskcommentservice "github.com/rijum8906/relay/services/task-service/internal/services/task_comment"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	logger, appErr := initLogger(a.config)
	if appErr != nil {
		return appErr
	}
	a.utils.logger = logger

	return nil
}

func (a *Application) initServices() *apperror.AppError {
	queries := taskdb.New(a.infra.database)
	projectRepository := projectRepo.NewProjectRepository(queries)
	projectMembershipRepository := projectmembershiprepo.NewProjectMembershipRepository(queries)
	taskRepository := taskrepo.NewTaskRepository(queries)
	taskAssignmentRepository := taskassignmentrepo.NewTaskAssignmentRepository(queries)
	taskCommentRepository := taskcommentrepo.NewTaskCommentRepository(queries)

	projectService, appErr := projectservice.NewProjectService(projectRepository)
	if appErr != nil {
		return appErr
	}
	a.services.project = projectService

	projectMembershipService, appErr := projectmembershipservice.NewProjectMembershipService(projectMembershipRepository)
	if appErr != nil {
		return appErr
	}
	a.services.projectMembership = projectMembershipService

	taskService, appErr := taskservice.NewTaskService(taskRepository)
	if appErr != nil {
		return appErr
	}
	a.services.task = taskService

	taskAssignmentService, appErr := taskassigmentservice.NewTaskAssignmentService(taskAssignmentRepository)
	if appErr != nil {
		return appErr
	}
	a.services.taskAssignment = taskAssignmentService

	taskCommentService, appErr := taskcommentservice.NewTaskCommentService(taskCommentRepository)
	if appErr != nil {
		return appErr
	}
	a.services.taskComment = taskCommentService
	return nil
}

func (a *Application) initHandler() *apperror.AppError {
	a.services.taskHandler = handler.NewTaskHandler(
		a.services.project,
		a.services.projectMembership,
		a.services.task,
		a.services.taskAssignment,
		a.services.taskComment,
	)
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
	a.server = server

	if a.config.AppEnv == "development" {
		reflection.Register(server)
	}

	taskv1.RegisterTaskServiceServer(server, a.services.taskHandler)

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

func (a *Application) Shutdown() {
	if a.server != nil {
		a.server.GracefulStop()
	}

	if a.infra != nil && a.infra.database != nil {
		a.infra.database.Close()
	}

	if a.infra != nil && a.infra.cache != nil {
		a.infra.cache.Close()
	}

	if a.infra != nil && a.infra.brokerClient != nil {
		_ = a.infra.brokerClient.Drain()
	}
}

func (a *Application) GetLogger() *zap.Logger {
	return a.utils.logger
}

func (a *Application) GetConfig() *config.Env {
	return a.config
}
