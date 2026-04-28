package organization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

func (s *organizationService) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest) (*modelsv1.Organization, error) {
	// Step 0. Validation
	if err := uuid.Validate(req.CreatedBy); err != nil {
		return nil, apperror.New(apperror.CodeValidation, "invalid user id")
	}
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, apperror.New(apperror.CodeValidation, "invalid create_by id")
	}
	isValidSlug := utils.ValidateSlug(req.Slug)
	if !isValidSlug {
		return nil, apperror.New(apperror.CodeValidation, "invalid slug")
	}

	// Step 1. Check user existstence and slug availability
	res, err := s.userClient.CheckExists(ctx, &userv1.CheckExistsRequest{
		Id: req.CreatedBy,
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, err.Error())
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
			return nil, apperror.New(apperror.CodeInternal, "error getting information from database").WithDetail("error", err.Error())
		}
	}

	// Step 2. Create Organization
	org, err := s.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: pgtype.Text{String: req.Description, Valid: true},
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// TODO: Step 3. Add user as owner in membership
	_, err = s.q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         createdBy,
		OrganizationID: org.ID,
		Role:           "owner",
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// TODO: Step 4. Add info to audit log

	// TODO: Step 5. Containerize everything in a db transaction

	return utils.MapOrganization(&org), nil
}

func (s *organizationService) GetOrganization(ctx context.Context, req *corev1.IDRequest) (*modelsv1.Organization, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.New(apperror.CodeValidation, "invalid organization id").WithDetail("error", err.Error())
	}

	org, err := s.q.GetOrganization(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to get organization info from database").WithDetail("error", err.Error())
	}

	return utils.MapOrganization(&org), nil
}

func (s *organizationService) GetOrganizationBySlug(ctx context.Context, req *organizationv1.GetOrganizationBySlugRequest) (*modelsv1.Organization, error) {
	// Validation
	isValidSlug := utils.ValidateSlug(req.Slug)
	if !isValidSlug {
		return nil, apperror.New(apperror.CodeValidation, "invalid slug")
	}

	org, err := s.q.GetOrganizationBySlug(ctx, req.Slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to get organization info from database").WithDetail("error", err.Error())
	}

	return utils.MapOrganization(&org), nil
}

func (s *organizationService) GetOrganizationsListByCreatedBy(ctx context.Context, req *corev1.EmptyRequest) (*organizationv1.OrganizationsList, error) {
	// Retrive user info from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.New(apperror.CodeInternal, "failed to extract user info from context")
	}
	if userInfo.UserID == "" {
		return nil, apperror.New(apperror.CodeInternal, "failed to extract user id from context")
	}

	// Step 0. Validation
	createdBy, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.New(apperror.CodeValidation, "invalid user id")
	}

	orgs, err := s.q.GetOrganizationsByCreatedBy(ctx, db.GetOrganizationsByCreatedByParams{
		CreatedBy: createdBy,
		Limit:     100,
		Offset:    0,
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "error getting information from database").WithDetail("error", err.Error())
	}

	orgsList := []*organizationv1.OrganizationResponse{}
	for _, org := range orgs {
		orgsList = append(orgsList, utils.MapOrganizationInfo(&org))
	}

	return &organizationv1.OrganizationsList{
		Organizations: orgsList,
	}, nil
}

func (s *organizationService) UpdateOrganizationName(ctx context.Context, req *organizationv1.UpdateOrganizationNameRequest) (*modelsv1.Organization, error) {
	// Step 0. Validation
	if err := uuid.Validate(req.OrganizationId); err != nil {
		return nil, apperror.New(apperror.CodeValidation, "invalid organization id")
	}
	isValid := token.ValidateTokenScope(req.TokenScope)
	if !isValid {
		return nil, apperror.New(apperror.CodeValidation, "invalid token scope")
	}

	// Step 1. Check Token Scope
	if req.TokenScope != string(token.TokenScopeUpdateOrganizationName) {
		return nil, apperror.New(apperror.CodeValidation, "invalid token scope")
	}

	// Step 2. Update
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	org, err := s.q.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		ID:          orgID,
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "couldn't update organization").WithDetail("error", err.Error())
	}

	return utils.MapOrganization(&org), nil
}

func (s *organizationService) ChangeOrganizationOwnership(context.Context, *organizationv1.ChangeOrganizationOwnershipRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *organizationService) DeleteOrganization(context.Context, *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *organizationService) ArchiveOrganization(context.Context, *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
