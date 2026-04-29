package taskcomment

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

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
