package authz

import (
	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
)

func UserObject(id uuid.UUID) string {
	return "user:" + id.String()
}

func ProjectObject(id uuid.UUID) string {
	return "project:" + id.String()
}

func TaskObject(id uuid.UUID) string {
	return "task:" + id.String()
}

func CommentObject(id uuid.UUID) string {
	return "task_comment:" + id.String()
}

func TeamObject(id uuid.UUID) string {
	return "team:" + id.String()
}

func ProjectRoleTuple(projectID uuid.UUID, role string, userID uuid.UUID) client.ClientTupleKey {
	return client.ClientTupleKey{
		User:     UserObject(userID),
		Relation: role,
		Object:   ProjectObject(projectID),
	}
}

func TaskCreatorTuple(taskID uuid.UUID, userID uuid.UUID) client.ClientTupleKey {
	return client.ClientTupleKey{
		User:     UserObject(userID),
		Relation: "creator",
		Object:   TaskObject(taskID),
	}
}

func TaskProjectTuple(taskID uuid.UUID, projectID uuid.UUID) client.ClientTupleKey {
	return client.ClientTupleKey{
		User:     ProjectObject(projectID),
		Relation: "parent_project",
		Object:   TaskObject(taskID),
	}
}

func TaskAssigneeTuple(taskID uuid.UUID, assigneeType string, assigneeID uuid.UUID) client.ClientTupleKey {
	assignee := UserObject(assigneeID)
	if assigneeType == "team" {
		assignee = TeamObject(assigneeID)
	}

	return client.ClientTupleKey{
		User:     assignee,
		Relation: "assignee",
		Object:   TaskObject(taskID),
	}
}

func CommentAuthorTuple(commentID uuid.UUID, userID uuid.UUID) client.ClientTupleKey {
	return client.ClientTupleKey{
		User:     UserObject(userID),
		Relation: "author",
		Object:   CommentObject(commentID),
	}
}

func CommentTaskTuple(commentID uuid.UUID, taskID uuid.UUID) client.ClientTupleKey {
	return client.ClientTupleKey{
		User:     TaskObject(taskID),
		Relation: "parent_task",
		Object:   CommentObject(commentID),
	}
}

func DeleteTuple(tuple client.ClientTupleKey) client.ClientTupleKeyWithoutCondition {
	return client.ClientTupleKeyWithoutCondition{
		User:     tuple.User,
		Relation: tuple.Relation,
		Object:   tuple.Object,
	}
}
