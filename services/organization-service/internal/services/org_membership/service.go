package orgmembership

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/protoutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

// ############################ USER QUERIES ############################

func (s *orgMembershipService) GetMyMemberships(ctx context.Context, req *corev1.PaginationRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate Pagination
	if appErr := protoutils.ValidatePaginationReq(req); appErr != nil {
		return nil, appErr
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithDetail("reason", "missing user metadata")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "invalid user uuid in context")
	}

	offset := (req.Page - 1) * req.Limit

	// 2. Retrieve Memberships from Database
	memberships, err := s.q.GetOrganizationMembershipsByUserID(ctx, db.GetOrganizationMembershipsByUserIDParams{
		UserID: userID,
		Limit:  req.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 4. Transform and Map Response
	// Pre-allocating the slice capacity improves performance by avoiding multiple re-allocations
	response := make([]*org_membershipv1.OrgMembershipRes, 0, len(memberships))
	for _, m := range memberships {
		response = append(response, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: response,
	}, nil
}

func (s *orgMembershipService) GetMyMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithDetail("reason", "missing user metadata")
	}

	membershipUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithDetail("id", "invalid membership uuid format")
	}

	// 2. Retrieve Membership
	// We fetch by both ID and UserID to enforce ownership at the database level.
	// This prevents ID-guessing attacks (IDOR).
	membership, err := s.q.GetOrganizationMembership(ctx, membershipUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithDetail("resource", "membership")
		}
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// 3. Security Authorization Check
	// Ensure the membership record actually belongs to the authenticated user.
	if membership.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrPermissionDenied.WithDetail("reason", "membership ownership mismatch")
	}

	// 4. Response Mapping
	return utils.MapOrgMembershipRes(&membership), nil
}

// ############################ ORGANIZATION QUERIES ############################

func (s *orgMembershipService) GetOrganizationMembershipsByOrgID(ctx context.Context, req *corev1.IDWithPaginationReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate Pagination
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}
	orgID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("organization not found")
	}

	// 3. Check permission via openfga
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + req.Id,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view this organization")
	}

	// 4. Get the organization memberships
	memberships, err := s.q.GetOrganizationMembershipsByOrgID(ctx, db.GetOrganizationMembershipsByOrgIDParams{
		OrganizationID: orgID,
		Limit:          req.Pagination.Limit,
		Offset:         (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Parse result
	result := []*org_membershipv1.OrgMembershipRes{}
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

func (s *orgMembershipService) GetOrganizationMembershipsByRole(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByRoleReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate Pagination
	if !permissions.ValidateRole(req.Role) {
		return nil, apperror.ErrValidation.WithMessage("invalid role")
	}
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("organization not found")
	}

	// 3. Check permission via openfga
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view this organization")
	}

	// 4. Get the organization memberships
	memberships, err := s.q.GetOrganizationMembershipsByOrgIDAndRole(ctx, db.GetOrganizationMembershipsByOrgIDAndRoleParams{
		OrganizationID: orgID,
		Role:           req.Role,
		Limit:          req.Pagination.Limit,
		Offset:         (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Parse result
	result := []*org_membershipv1.OrgMembershipRes{}
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

func (s *orgMembershipService) GetOrganizationMembershipsByStatus(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByStatusReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate Pagination
	if !utils.ValidateOrgnaziationMembershipStatus(req.Status) {
		return nil, apperror.ErrValidation.WithMessage("invalid status")
	}
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Check if organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("organization not found")
	}

	// 3. Check permission via openfga
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view this organization")
	}

	// 4. Get the organization memberships
	memberships, err := s.q.GetOrganizationMembershipsByOrgIDAndStatus(ctx, db.GetOrganizationMembershipsByOrgIDAndStatusParams{
		OrganizationID: orgID,
		Status:         req.Status,
		Limit:          req.Pagination.Limit,
		Offset:         (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Parse result
	result := []*org_membershipv1.OrgMembershipRes{}
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

func (s *orgMembershipService) GetOrganizationMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	// 0. Validate Pagination
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("invalid request body")
	}
	orgMemID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Get the membership
	membership, err := s.q.GetOrganizationMembership(ctx, orgMemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("membership not found")
		}
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch membership").WithDetail("db_error", err.Error())
	}

	// 3. Check permission via openfga
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + membership.OrganizationID.String(),
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view this organization")
	}

	// 4. Parse and return result
	return utils.MapOrgMembershipRes(&membership), nil
}

// ############################ INVITATION FLOW ############################

func (s *orgMembershipService) SendInvitation(ctx context.Context, req *org_membershipv1.SendInvitationRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request
	if appErr := protoutils.ValidateSendInvitationReq(req); appErr != nil {
		return nil, appErr
	}
	emailReq := &corev1.EmailRequest{
		Email: req.Email,
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}
	inviteBy, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("reason", "failed to parse user id").WithDetail("error", err.Error())
	}

	// 2. Check if email exists
	exits, err := s.userClient.CheckEmailExists(ctx, emailReq)
	if err != nil {
		return nil, apperror.ErrThirdParty.WithDetail("error", err.Error())
	}
	if !exits.Exists {
		return nil, apperror.ErrNotFound.WithMessage("email not found")
	}

	// 3. Check If he has permission to send invitation
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + inviteBy.String(),
		Relation: permissions.PermissionCanAddMember,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to invite members")
	}

	// 4. Get This user's membership id
	membership, err := s.q.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
		UserID:         inviteBy,
		OrganizationID: uuid.MustParse(req.OrganizationId),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch membership").WithDetail("db_error", err.Error())
	}

	// 5. Create an invitation
	tokenHash, appErr := s.hashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}
	_, err = s.q.CreateOrganizationInvitation(ctx, db.CreateOrganizationInvitationParams{
		Email:          req.Email,
		OrganizationID: uuid.MustParse(req.OrganizationId),
		Role:           req.Role,
		InvitedByMemID: membership.ID,
		TokenHash:      tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * time.Duration(s.config.InvitationTokenTTL)),
			Valid: true,
		},
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to create invitation").WithDetail("db_error", err.Error())
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *orgMembershipService) AcceptInvitation(ctx context.Context, req *corev1.TokenHashRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}
	if req.TokenHash == "" {
		return nil, apperror.ErrValidation.WithMessage("token hash cannot be empty")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Fetch user details from user service to get email for validation
	user, err := s.userClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	// 2. Retrieve invitation using the provided token hash
	invitation, err := s.q.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("invitation not found or already expired")
		}
		return nil, apperror.ErrInternal.WithMessage("failed to fetch invitation").WithDetail("db_error", err.Error())
	}

	// 3. Validate invitation is still valid
	// Check if invitation has expired
	if time.Now().After(invitation.ExpiresAt.Time) {
		return nil, apperror.ErrValidation.WithMessage("invitation has expired")
	}
	// Check if invitation is still in pending state (not accepted or cancelled)
	if invitation.Status != "pending" {
		return nil, apperror.ErrNotFound.WithMessage("invitation not found")
	}

	// 4. Verify the authenticated user's email matches the invitation recipient
	if invitation.Email != user.Email {
		return nil, apperror.ErrPermissionDenied.WithMessage("this invitation was sent to a different email address")
	}

	// 5. Create organization membership for the user
	_, err = s.q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		UserID:         uuid.MustParse(userInfo.UserID),
		OrganizationID: invitation.OrganizationID,
		Role:           invitation.Role,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create organization membership").WithDetail("db_error", err.Error())
	}

	// 6. Add user permissions to OpenFGA
	if appErr := s.tuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleMember,
			Object:   "organization:" + invitation.OrganizationID.String(),
		},
	}); appErr != nil {
		return nil, appErr
	}

	// 7. Mark the invitation as accepted
	_, err = s.q.AccecptOrganizationInvitation(ctx, db.AccecptOrganizationInvitationParams{
		ID:          invitation.ID,
		RespondedBy: uuid.MustParse(userInfo.UserID),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to accept invitation").WithDetail("db_error", err.Error())
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *orgMembershipService) DeclineInvitation(ctx context.Context, req *corev1.TokenHashRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}
	if req.TokenHash == "" {
		return nil, apperror.ErrValidation.WithMessage("token hash cannot be empty")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Fetch user details from user service to get email for validation
	user, err := s.userClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	// 2. Retrieve invitation using the provided token hash
	invitation, err := s.q.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("invitation not found or already expired")
		}
		return nil, apperror.ErrInternal.WithMessage("failed to fetch invitation").WithDetail("db_error", err.Error())
	}

	// 3. Validate invitation is still valid
	// Check if invitation has expired
	if time.Now().After(invitation.ExpiresAt.Time) {
		return nil, apperror.ErrValidation.WithMessage("invitation has expired")
	}
	// Check if invitation is still in pending state (not accepted or cancelled)
	if invitation.Status != "pending" {
		return nil, apperror.ErrNotFound.WithMessage("invitation not found")
	}

	// 4. Verify the authenticated user's email matches the invitation recipient
	if invitation.Email != user.Email {
		return nil, apperror.ErrPermissionDenied.WithMessage("this invitation was sent to a different email address")
	}

	// 7. Mark the invitation as rejected
	_, err = s.q.DeclineOrganizationInvitation(ctx, db.DeclineOrganizationInvitationParams{
		ID:          invitation.ID,
		RespondedBy: uuid.MustParse(userInfo.UserID),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to accept invitation").WithDetail("db_error", err.Error())
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// ############################ USER'S MEMBERSHIP MANAGEMENT ############################

func (s *orgMembershipService) LeaveOrganization(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

// ############################ ORGANIZATION'S MEMBERSHIP MANAGEMENT ############################

func (s *orgMembershipService) ChangeOrganizationMembershipStatus(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipStatusReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) ChangeOrganizationMembershipRole(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipRoleReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) RemoveOrganizationMember(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
