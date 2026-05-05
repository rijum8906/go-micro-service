package organization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/protoutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

// CreateOrganization
//   - Creates an organization
//   - Then creates a organizatio member for the user and make him the owner
//   - Added ownership to openfga
func (s *organizationService) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest) (*modelsv1.Organization, error) {
	// Step 0. Validation
	if appErr := validateCreateOrganizationRequest(req); appErr != nil {
		return nil, appErr
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

	exists, err := s.q.CheckOrganizationExistsBySlug(ctx, req.Slug)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if exists {
		return nil, apperror.New(apperror.CodeValidation, "slug already exists")
	}

	// Step 2. Create Organization
	createdBy, _ := uuid.Parse(req.CreatedBy)
	org, err := s.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: pgtype.Text{String: req.Description, Valid: true},
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// Step 3. Add user as owner in membership
	_, err = s.q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         createdBy,
		OrganizationID: org.ID,
		Role:           "owner",
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// Step 4. Add relation to openFgaClient
	if appErr := s.tuppleManager.Write(ctx, []client.ClientTupleKey{{
		User:     "user:" + req.CreatedBy,
		Relation: permissions.RoleOwner,
		Object:   "organization:" + org.ID.String(),
	}}); appErr != nil {
		return nil, appErr
	}

	// TODO: Step 4. Add info to audit log

	// TODO: Step 5. Containerize everything in a db transaction

	return utils.MapOrganization(&org), nil
}

func (s *organizationService) GetOrganization(ctx context.Context, req *corev1.IDRequest) (*modelsv1.Organization, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid id")
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
	if err := validateUpdateOrganizationName(req); err != nil {
		return nil, err
	}

	// Step 1. Extract User Info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user info from context")
	}
	if userInfo.UserID == "" {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user id from context")
	}

	orgID, _ := uuid.Parse(req.OrganizationId)

	// Step 1. Check Token Scope
	if req.TokenScope != string(token.TokenScopeUpdateOrganizationName) {
		return nil, apperror.New(apperror.CodeValidation, "invalid token scope")
	}

	// Step 2. Check OpenFGA permission
	res, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanEdit,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*res.Allowed {
		return nil, apperror.New(apperror.CodeForbidden, "user is not allowed to update organization name")
	}

	// Step 3. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeNotFound, "organization does not exist")
	}

	// Step 4. Update
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

func (s *organizationService) ChangeOrganizationOwnership(ctx context.Context, req *organizationv1.ChangeOrganizationOwnershipRequest) (*corev1.SuccessResponse, error) {
	// Step 0. Validation
	if appErr := validateChangeOwnershipRequst(req); appErr != nil {
		return nil, appErr
	}

	orgID, _ := uuid.Parse(req.OrganizationId)
	newOwnerID, _ := uuid.Parse(req.NewOwnerId)

	// Step 1. Extract User Info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user info from context")
	}
	if userInfo.UserID == "" {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user id from context")
	}

	// Step 2. Check if user exists
	res, err := s.userClient.CheckExists(ctx, &userv1.CheckExistsRequest{
		Id: req.NewOwnerId,
	})
	if err != nil {
		return nil, apperror.ErrThirdParty.WithDetail("error", err.Error())
	}
	if !res.Exists {
		return nil, apperror.New(apperror.CodeNotFound, "user does not exist")
	}

	// Step 3. Check OpenFGA permission
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanDelete,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.New(apperror.CodeForbidden, "user is not allowed to change organization ownership")
	}

	// Step 4. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeNotFound, "organization does not exist")
	}

	// Step 5. Check Token Scope
	if req.TokenScope != string(token.TokenScopeChangeOrganizationOwner) {
		return nil, apperror.New(apperror.CodeValidation, "invalid token scope")
	}

	// Step 6. Change Ownership
	err = s.q.ChangeOrganizationOwnership(ctx, db.ChangeOrganizationOwnershipParams{
		ID:        orgID,
		CreatedBy: newOwnerID,
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "couldn't change organization ownership").WithDetail("error", err.Error())
	}

	// Step 7. Update Fga client owner ship
	if appErr = s.tuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleOwner,
			Object:   "organization:" + req.OrganizationId,
		},
	}); appErr != nil {
		return nil, appErr
	}
	if appErr := s.tuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + newOwnerID.String(),
			Relation: permissions.RoleOwner,
			Object:   "organization:" + req.OrganizationId,
		},
	}); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// DeleteOrganization Soft deletes a organization
// NOTE: Gateway validates organization ownership before issuing scoped token.
// However, we still check - Token Scope, Organization Existence
// but not the stuffs that comes from the token (token can't be altered by any means)
func (s *organizationService) DeleteOrganization(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	// Step 0. Validation
	if err := protoutils.ValidateIDAndScopedToken(req); err != nil {
		return nil, err
	}

	orgID, _ := uuid.Parse(req.Id)

	// Step 1. Extract User Info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user info from context")
	}
	if userInfo.UserID == "" {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user id from context")
	}

	deletedBy, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to parse user id").WithDetail("error", err.Error())
	}

	// Step 2. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeNotFound, "organization does not exist")
	}

	// Step 3. Check Token Scope
	if req.TokenScope != string(token.TokenScopeDeleteOrganization) {
		return nil, apperror.New(apperror.CodeValidation, "invalid token scope")
	}

	// Step 4. Check OpenFGA permission
	res, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanDelete,
		Object:   "organization:" + req.Id,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*res.Allowed {
		return nil, apperror.New(apperror.CodeForbidden, "user is not allowed to delete organization")
	}

	// Step 5. Delete
	err = s.q.DeleteOrganization(ctx, db.DeleteOrganizationParams{
		ID:        orgID,
		DeletedBy: deletedBy,
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "couldn't delete organization").WithDetail("error", err.Error())
	}

	if appErr = s.tuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleOwner,
			Object:   "organization:" + req.Id,
		},
	}); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *organizationService) ArchiveOrganization(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	// Step 0. Validation
	if err := protoutils.ValidateIDAndScopedToken(req); err != nil {
		return nil, err
	}

	orgID, _ := uuid.Parse(req.Id)

	// Step 1. Extract User Info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user info from context")
	}
	if userInfo.UserID == "" {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user id from context")
	}

	// Step 2. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeNotFound, "organization does not exist")
	}

	// Step 3. Check Token Scope
	if req.TokenScope != string(token.TokenScopeArchiveOrganization) {
		return nil, apperror.New(apperror.CodeValidation, "invalid token scope")
	}

	// Step 4. Check OpenFGA permission
	res, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanDelete,
		Object:   "organization:" + req.Id,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*res.Allowed {
		return nil, apperror.New(apperror.CodeForbidden, "user is not allowed to archive organization")
	}

	// Step 5. Archive
	err = s.q.ArchiveOrganization(ctx, orgID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "couldn't archive organization").WithDetail("error", err.Error())
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
