package orgteam

import (
	"context"

	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_teamv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_team/v1"
)

func (s *OrgTeamService) ValidateTeamAccess(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *OrgTeamService) GetTeamPermissions(ctx context.Context, req *corev1.IDRequest) (*org_teamv1.TeamPermissionsRes, error) {
	return nil, nil
}
