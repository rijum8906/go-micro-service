package orgteam

import (
	"context"

	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	org_teamv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_team/v1"
)

func (s *OrgTeamService) GetOrganizationTeamsByOrgID(ctx context.Context, req *corev1.IDWithPaginationReq) (*org_teamv1.OrgTeamListRes, error) {
	return nil, nil
}

func (s *OrgTeamService) GetOrganizationTeam(ctx context.Context, req *corev1.IDRequest) (*org_teamv1.OrgTeamRes, error) {
	return nil, nil
}

func (s *OrgTeamService) GeOrganizationTeamWithMeta(ctx context.Context, req *corev1.IDRequest) (*modelsv1.OrganizationTeam, error) {
	return nil, nil
}
