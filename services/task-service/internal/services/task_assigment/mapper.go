package taskassigment

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func mapTaskAssignment(assignment *db.TaskAssignment) *modelsv1.TaskAssignment {
	if assignment == nil {
		return nil
	}

	return &modelsv1.TaskAssignment{
		Id:           assignment.ID.String(),
		TaskId:       assignment.TaskID.String(),
		AssigneeType: assignment.AssigneeType,
		AssigneeId:   assignment.AssigneeID.String(),
		AssignedBy:   assignment.AssignedBy.String(),
		AssignedAt:   utils.Timestamp(assignment.AssignedAt),
		UnassignedAt: utils.Timestamp(assignment.UnassignedAt),
		CreatedAt:    utils.Timestamp(assignment.CreatedAt),
		UpdatedAt:    utils.Timestamp(assignment.UpdatedAt),
	}
}

func mapTaskAssignments(assignments []db.TaskAssignment) []*modelsv1.TaskAssignment {
	items := make([]*modelsv1.TaskAssignment, 0, len(assignments))
	for i := range assignments {
		items = append(items, mapTaskAssignment(&assignments[i]))
	}

	return items
}
