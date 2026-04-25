package organization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest) (*modelsv1.Organization, error) {
	_, err := s.q.GetOrganizationBySlug(ctx, req.Slug)
	if err == nil {
		return nil, errors.New("organization already exists with slug " + req.Slug)
	} else {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("internal server error" + err.Error())
		}
	}

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, errors.New("invalid created_by")
	}

	org, err := s.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: pgtype.Text{String: req.Description, Valid: true},
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, errors.New("internal server error" + err.Error())
	}

	return utils.MapOrganization(&org), nil
}
