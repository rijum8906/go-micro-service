package orgteam

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_teamv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_team/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func (s *OrgTeamService) CreateTeam(ctx context.Context, req *org_teamv1.CreateOrgTeamRequest) (*org_teamv1.OrgTeamRes, error) {
	// Validate request parameters
	orgID, membershipID, appErr := parseCreateOrgTeamReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}
	userID, _ := uuid.Parse(userInfo.UserID)

	// Check if user has permission to create a team
	res, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanCreateTeam,
		Object:   "object:" + orgID.String(),
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*res.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("you do not have permission to create a team")
	}

	// Check if organization team name already exists
	exists, err := s.DBQ.CheckOrganizationTeamNameExists(ctx, db.CheckOrganizationTeamNameExistsParams{
		Name:           req.Name,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}
	if exists {
		return nil, apperror.New(apperror.CodeValidation, "team name already exists")
	}

	var team *db.OrganizationTeam
	var teamMembership *db.OrganizationTeamMembership

	s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Create t
		t, err := q.CreateOrganizationTeam(ctx, db.CreateOrganizationTeamParams{
			Description: pgtype.Text{
				String: req.Description,
				Valid:  req.Description != "",
			},
			Name:           req.Name,
			OrganizationID: orgID,
			CreatedByMemID: userID,
		})
		if err != nil {
			return apperror.ErrInternal.WithDetail("error", err.Error())
		}
		team = &t

		// create team membership and make this user admin
		tm, err := q.CreateOrganizationTeamMembership(ctx, db.CreateOrganizationTeamMembershipParams{
			OrganizationID: orgID,
			Role:           constants.OrgTeamRoleOwner,
			TeamID:         t.ID,
			MembershipID:   membershipID,
		})
		if err != nil {
			return apperror.ErrInternal.WithDetail("error", err.Error())
		}
		teamMembership = &tm
		return nil
	})

	// create role
	if appErr = s.Helper.AddTeamMemRole(ctx, teamMembership); appErr != nil {
		return nil, appErr
	}

	return &org_teamv1.OrgTeamRes{
		Id:             team.ID.String(),
		Name:           team.Name,
		Description:    team.Description.String,
		OrganizationId: team.OrganizationID.String(),
	}, nil
}

func (s *OrgTeamService) UpdateTeamName(ctx context.Context, req *org_teamv1.UpdateOrgTeamNameRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *OrgTeamService) DeleteTeam(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
