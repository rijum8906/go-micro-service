package orgteam

import (
	"context"

	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_teamv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_team/v1"
)

func (s *OrgTeamService) CreateTeam(ctx context.Context, req *org_teamv1.CreateOrgTeamRequest) (*org_teamv1.OrgTeamRes, error) {
	return nil, nil
}

func (s *OrgTeamService) UpdateTeamName(ctx context.Context, req *org_teamv1.UpdateOrgTeamNameRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *OrgTeamService) DeleteTeam(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
