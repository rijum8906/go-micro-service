package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func (s *service) CreateOrganization(ctx context.Context, name, slug, desc string, createdBy uuid.UUID) (*db.Organization, *apperror.AppError) {
	return nil, nil
}
