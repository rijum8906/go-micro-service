package organization

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func (s *service) CreateOrganization(ctx context.Context, name, description string, createdBy uuid.UUID) (*db.Organization, *apperror.AppError) {
	org, err := s.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:        name,
		Description: pgtype.Text{String: description, Valid: true},
		Slug:        strings.ToLower(name),
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	return &org, nil
}
