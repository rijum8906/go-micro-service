package organization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
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
	_, err = s.q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         createdBy,
		OrganizationID: org.ID,
		Role:           "owner",
	})
	if err != nil {
		return nil, errors.New("failed to create organization" + err.Error())
	}

	// TODO: Step 4. Add info to audit log

	// TODO: Step 5. Containerize everything in a db transaction

	return utils.MapOrganization(&org), nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, req *corev1.IDRequest) (*modelsv1.Organization, error) {
	return nil, nil
}

func (s *OrganizationService) GetOrganizationBySlug(ctx context.Context, req *organizationv1.GetOrganizationBySlugRequest) (*modelsv1.Organization, error) {
	return nil, nil
}

func (s *OrganizationService) GetOrganizationsListByCreatedBy(ctx context.Context, req *corev1.EmptyRequest) (*organizationv1.OrganizationsList, error) {
	return nil, nil
}

func (s *OrganizationService) UpdateOrganizationName(context.Context, *organizationv1.UpdateOrganizationNameRequest) (*modelsv1.Organization, error) {
	return nil, nil
}

func (s *OrganizationService) ChangeOrganizationOwnership(context.Context, *corev1.IDRequest) (*modelsv1.Organization, error) {
	return nil, nil
}

func (s *OrganizationService) DeleteOrganization(context.Context, *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *OrganizationService) ArchiveOrganization(context.Context, *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
