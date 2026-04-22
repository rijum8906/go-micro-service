package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

type Service interface {
	CreateOrganization(ctx context.Context, name, desc string, createdBy uuid.UUID) (*db.Organization, *apperror.AppError)
}

type service struct {
	q db.Querier
}

func New(q db.Querier) Service {
	return &service{
		q: q,
	}
}
