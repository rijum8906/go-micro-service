// Package organization service
package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"google.golang.org/grpc"
)

type Service interface {
	CreateOrganization(ctx context.Context, name, slug, desc string, createdBy uuid.UUID) (*db.Organization, *apperror.AppError)
}

type service struct {
	q          db.Querier
	grpcServer *grpc.Server
}

func New(q db.Querier, client *userv1.UserServiceClient) Service {
	grpcUserClient := userv1.NewUserServiceClient(client)
	return &service{
		q: q,
	}
}
