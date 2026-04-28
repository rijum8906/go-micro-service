package project

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func mapProject(project *db.Project) *modelsv1.Project {
	if project == nil {
		return nil
	}

	return &modelsv1.Project{
		Id:             project.ID.String(),
		OrganizationId: utils.UUIDString(project.OrganizationID),
		CreatedBy:      project.CreatedBy.String(),
		Name:           project.Name,
		Description:    project.Description,
		Status:         project.Status,
		ArchivedAt:     utils.Timestamp(project.ArchivedAt),
		CompletedAt:    utils.Timestamp(project.CompletedAt),
		DeletedAt:      utils.Timestamp(project.DeletedAt),
		DeletedBy:      utils.UUIDString(project.DeletedBy),
		CreatedAt:      utils.Timestamp(project.CreatedAt),
		UpdatedAt:      utils.Timestamp(project.UpdatedAt),
	}
}

func mapProjects(projects []db.Project) []*modelsv1.Project {
	items := make([]*modelsv1.Project, 0, len(projects))
	for i := range projects {
		items = append(items, mapProject(&projects[i]))
	}

	return items
}
