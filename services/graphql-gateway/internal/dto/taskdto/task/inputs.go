package taskdto

type CreateTaskInput struct {
	Title          string  `json:"title" validate:"required"`
	OrganizationID *string `json:"organizationId" validate:"omitempty,uuid4"`
	ProjectID      *string `json:"projectId" validate:"omitempty,uuid4"`
	ParentTaskID   *string `json:"parentTaskId" validate:"omitempty,uuid4"`
	Description    *string `json:"description"`
	Priority       *string `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueAt          *string `json:"dueAt" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type GetTaskInput struct {
	ID string `json:"id" validate:"required,uuid4"`
}

type ListTasksByProjectInput struct {
	ProjectID string `json:"projectId" validate:"required,uuid4"`
}
