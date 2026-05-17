package task

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
)

func (s *service) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*modelsv1.Task, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
		return s.createTask(ctx, req, userInfo)
	})
}

func (s *service) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
		return s.getTask(ctx, req, userInfo)
	})
}

func (s *service) ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListTasksByProjectResponse, *apperror.AppError) {
		return s.listTasksByProject(ctx, req, userInfo)
	})
}

func (s *service) CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest) (*modelsv1.Project, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
		return s.createProject(ctx, req, userInfo)
	})
}

func (s *service) GetProject(ctx context.Context, req *taskv1.GetProjectRequest) (*modelsv1.Project, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
		return s.getProject(ctx, req, userInfo)
	})
}

func (s *service) UpdateProject(ctx context.Context, req *taskv1.UpdateProjectRequest) (*modelsv1.Project, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
		return s.updateProject(ctx, req, userInfo)
	})
}

func (s *service) CompleteProject(ctx context.Context, req *taskv1.CompleteProjectRequest) (*modelsv1.Project, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
		return s.completeProject(ctx, req, userInfo)
	})
}

func (s *service) ArchiveProject(ctx context.Context, req *taskv1.ArchiveProjectRequest) (*modelsv1.Project, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Project, *apperror.AppError) {
		return s.archiveProject(ctx, req, userInfo)
	})
}

func (s *service) DeleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest) (*corev1.SuccessResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
		return s.deleteProject(ctx, req, userInfo)
	})
}

func (s *service) ListProjects(ctx context.Context, req *taskv1.ListProjectsRequest) (*taskv1.ListProjectsResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListProjectsResponse, *apperror.AppError) {
		return s.listProjects(ctx, req, userInfo)
	})
}

func (s *service) AddProjectMember(ctx context.Context, req *taskv1.AddProjectMemberRequest) (*modelsv1.ProjectMembership, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError) {
		return s.addProjectMember(ctx, req, userInfo)
	})
}

func (s *service) RemoveProjectMember(ctx context.Context, req *taskv1.RemoveProjectMemberRequest) (*corev1.SuccessResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
		return s.removeProjectMember(ctx, req, userInfo)
	})
}

func (s *service) UpdateProjectMemberRole(ctx context.Context, req *taskv1.UpdateProjectMemberRoleRequest) (*modelsv1.ProjectMembership, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError) {
		return s.updateProjectMemberRole(ctx, req, userInfo)
	})
}

func (s *service) ListProjectMembers(ctx context.Context, req *taskv1.ListProjectMembersRequest) (*taskv1.ListProjectMembersResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListProjectMembersResponse, *apperror.AppError) {
		return s.listProjectMembers(ctx, req, userInfo)
	})
}

func (s *service) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*modelsv1.Task, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
		return s.updateTask(ctx, req, userInfo)
	})
}

func (s *service) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*corev1.SuccessResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
		return s.deleteTask(ctx, req, userInfo)
	})
}

func (s *service) ArchiveTask(ctx context.Context, req *taskv1.ArchiveTaskRequest) (*modelsv1.Task, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
		return s.archiveTask(ctx, req, userInfo)
	})
}

func (s *service) UpdateTaskStatus(ctx context.Context, req *taskv1.UpdateTaskStatusRequest) (*modelsv1.Task, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
		return s.updateTaskStatus(ctx, req, userInfo)
	})
}

func (s *service) UpdateTaskProgress(ctx context.Context, req *taskv1.UpdateTaskProgressRequest) (*modelsv1.Task, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.Task, *apperror.AppError) {
		return s.updateTaskProgress(ctx, req, userInfo)
	})
}

func (s *service) ListTasksByOrganization(ctx context.Context, req *taskv1.ListTasksByOrganizationRequest) (*taskv1.ListTasksByOrganizationResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListTasksByOrganizationResponse, *apperror.AppError) {
		return s.listTasksByOrganization(ctx, req, userInfo)
	})
}

func (s *service) ListTasksByParent(ctx context.Context, req *taskv1.ListTasksByParentRequest) (*taskv1.ListTasksByParentResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListTasksByParentResponse, *apperror.AppError) {
		return s.listTasksByParent(ctx, req, userInfo)
	})
}

func (s *service) ListTasksByCreator(ctx context.Context, req *taskv1.ListTasksByCreatorRequest) (*taskv1.ListTasksByCreatorResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListTasksByCreatorResponse, *apperror.AppError) {
		return s.listTasksByCreator(ctx, req, userInfo)
	})
}

func (s *service) AssignTask(ctx context.Context, req *taskv1.AssignTaskRequest) (*modelsv1.TaskAssignment, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.TaskAssignment, *apperror.AppError) {
		return s.assignTask(ctx, req, userInfo)
	})
}

func (s *service) UnassignTask(ctx context.Context, req *taskv1.UnassignTaskRequest) (*corev1.SuccessResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
		return s.unassignTask(ctx, req, userInfo)
	})
}

func (s *service) ReassignTask(ctx context.Context, req *taskv1.ReassignTaskRequest) (*modelsv1.TaskAssignment, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.TaskAssignment, *apperror.AppError) {
		return s.reassignTask(ctx, req, userInfo)
	})
}

func (s *service) ListTaskAssignments(ctx context.Context, req *taskv1.ListTaskAssignmentsRequest) (*taskv1.ListTaskAssignmentsResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListTaskAssignmentsResponse, *apperror.AppError) {
		return s.listTaskAssignments(ctx, req, userInfo)
	})
}

func (s *service) CreateTaskComment(ctx context.Context, req *taskv1.CreateTaskCommentRequest) (*modelsv1.TaskComment, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError) {
		return s.createTaskComment(ctx, req, userInfo)
	})
}

func (s *service) UpdateTaskComment(ctx context.Context, req *taskv1.UpdateTaskCommentRequest) (*modelsv1.TaskComment, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*modelsv1.TaskComment, *apperror.AppError) {
		return s.updateTaskComment(ctx, req, userInfo)
	})
}

func (s *service) DeleteTaskComment(ctx context.Context, req *taskv1.DeleteTaskCommentRequest) (*corev1.SuccessResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
		return s.deleteTaskComment(ctx, req, userInfo)
	})
}

func (s *service) ListTaskComments(ctx context.Context, req *taskv1.ListTaskCommentsRequest) (*taskv1.ListTaskCommentsResponse, error) {
	return withUserInfo(ctx, func(userInfo *coredto.UserInfo) (*taskv1.ListTaskCommentsResponse, *apperror.AppError) {
		return s.listTaskComments(ctx, req, userInfo)
	})
}
