package task

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func mapTask(task *db.Task) *modelsv1.Task {
	if task == nil {
		return nil
	}

	return &modelsv1.Task{
		Id:              task.ID.String(),
		OrganizationId:  utils.UUIDString(task.OrganizationID),
		ProjectId:       utils.UUIDString(task.ProjectID),
		ParentTaskId:    utils.UUIDString(task.ParentTaskID),
		CreatedBy:       task.CreatedBy.String(),
		UpdatedBy:       utils.UUIDString(task.UpdatedBy),
		Title:           task.Title,
		Description:     task.Description,
		Status:          task.Status,
		Priority:        task.Priority,
		ProgressPercent: int32(task.ProgressPercent),
		StartedAt:       utils.Timestamp(task.StartedAt),
		DueAt:           utils.Timestamp(task.DueAt),
		CompletedAt:     utils.Timestamp(task.CompletedAt),
		ArchivedAt:      utils.Timestamp(task.ArchivedAt),
		DeletedAt:       utils.Timestamp(task.DeletedAt),
		DeletedBy:       utils.UUIDString(task.DeletedBy),
		CreatedAt:       utils.Timestamp(task.CreatedAt),
		UpdatedAt:       utils.Timestamp(task.UpdatedAt),
	}
}

func mapTasks(tasks []db.Task) []*modelsv1.Task {
	items := make([]*modelsv1.Task, 0, len(tasks))
	for i := range tasks {
		items = append(items, mapTask(&tasks[i]))
	}
	return items
}

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

func mapTaskComment(comment *db.TaskComment) *modelsv1.TaskComment {
	if comment == nil {
		return nil
	}

	return &modelsv1.TaskComment{
		Id:        comment.ID.String(),
		TaskId:    comment.TaskID.String(),
		AuthorId:  comment.AuthorID.String(),
		Body:      comment.Body,
		EditedAt:  utils.Timestamp(comment.EditedAt),
		EditCount: comment.EditCount,
		DeletedAt: utils.Timestamp(comment.DeletedAt),
		DeletedBy: utils.UUIDString(comment.DeletedBy),
		CreatedAt: utils.Timestamp(comment.CreatedAt),
		UpdatedAt: utils.Timestamp(comment.UpdatedAt),
	}
}

func mapTaskComments(comments []db.TaskComment) []*modelsv1.TaskComment {
	items := make([]*modelsv1.TaskComment, 0, len(comments))
	for i := range comments {
		items = append(items, mapTaskComment(&comments[i]))
	}
	return items
}
