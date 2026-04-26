package organization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest) (*modelsv1.Organization, error) {
	// TODO: Step 0. Validation
	// Step 1. Check user existstence and slug availability
	res, err := s.userClient.CheckExists(ctx, &userv1.CheckExistsRequest{
		Id: req.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	if !res.Exists {
		return nil, errors.New("user does not exist")
	}

	// WARN: use other method to validate slug (eg. bloom filter)
	_, err = s.q.GetOrganizationBySlug(ctx, req.Slug)
	if err == nil {
		return nil, errors.New("organization already exists with slug " + req.Slug)
	} else {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("internal server error" + err.Error())
		}
	}

	// Step 2. Create Organization
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

	// TODO: Step 3. Add user as owner in membership

	// TODO: Step 4. Add info to audit log

	// TODO: Step 5. Containerize everything in a db transaction

	return utils.MapOrganization(&org), nil
}
