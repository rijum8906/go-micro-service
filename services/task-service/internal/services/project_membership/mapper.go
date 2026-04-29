package projectmembership

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func mapProjectMembership(membership *db.ProjectMembership) *modelsv1.ProjectMembership {
	if membership == nil {
		return nil
	}

	return &modelsv1.ProjectMembership{
		Id:        membership.ID.String(),
		ProjectId: membership.ProjectID.String(),
		UserId:    membership.UserID.String(),
		Role:      membership.Role,
		JoinedAt:  utils.Timestamp(membership.JoinedAt),
		LeftAt:    utils.Timestamp(membership.LeftAt),
		CreatedAt: utils.Timestamp(membership.CreatedAt),
		UpdatedAt: utils.Timestamp(membership.UpdatedAt),
	}
}

func mapProjectMemberships(memberships []db.ProjectMembership) []*modelsv1.ProjectMembership {
	items := make([]*modelsv1.ProjectMembership, 0, len(memberships))
	for i := range memberships {
		items = append(items, mapProjectMembership(&memberships[i]))
	}

	return items
}
