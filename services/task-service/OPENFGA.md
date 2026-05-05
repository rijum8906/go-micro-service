# Task-Service OpenFGA Draft

This draft is based on the current authorization behavior in `task-service`, with a conservative v1 target for OpenFGA integration.

## Current Rules

### Project

| Action | Current rule |
| --- | --- |
| `CreateProject` | any authenticated user |
| `GetProject` | `member` or higher |
| `UpdateProject` | `admin` or higher |
| `CompleteProject` | `admin` or higher |
| `ArchiveProject` | `admin` or higher |
| `DeleteProject` | `owner` only |
| `AddProjectMember(member/admin)` | `admin` or higher |
| `AddProjectMember(owner)` | `owner` only |
| `RemoveProjectMember` | `admin` or higher |
| `UpdateProjectMemberRole` | `owner` only |
| `ListProjectMembers` | `member` or higher |

### Task

`RequireTaskRole` currently works like this:

- if the task belongs to a project, use the caller's role on that project
- if the task does not belong to a project, only the task creator has access
- for standalone tasks, the `minRole` argument is effectively ignored, so the creator can also delete, archive, and manage assignments

| Action | Current rule |
| --- | --- |
| `CreateTask` in project | `member` or higher on project |
| `CreateTask` standalone or org-level | any authenticated user |
| `GetTask` | project `member` or higher, or standalone creator |
| `ListTasksByProject` | project `member` or higher |
| `UpdateTask` | project `member` or higher, or standalone creator |
| `UpdateTaskStatus` | project `member` or higher, or standalone creator |
| `UpdateTaskProgress` | project `member` or higher, or standalone creator |
| `DeleteTask` | project `admin` or higher, or standalone creator |
| `ArchiveTask` | project `admin` or higher, or standalone creator |
| `ListTasksByParent` | access to parent task, then scope filtering |

### Task Assignment

| Action | Current rule |
| --- | --- |
| `AssignTask` | project `admin` or higher, or standalone creator |
| `UnassignTask` | project `admin` or higher, or standalone creator |
| `ReassignTask` | project `admin` or higher, or standalone creator |
| `ListTaskAssignments` | project `member` or higher, or standalone creator |

Note: current assignment records do not grant access by themselves. Being an assignee is stored, but not used for authorization yet.

### Task Comment

| Action | Current rule |
| --- | --- |
| `CreateTaskComment` | project `member` or higher, or standalone creator |
| `ListTaskComments` | project `member` or higher, or standalone creator |
| `UpdateTaskComment` | task access plus comment author only |
| `DeleteTaskComment` | task access plus comment author only |

## V1 Decisions

- Project creator becomes `owner` by writing an OpenFGA tuple after project creation.
- Project `member` can view project data, list members, create tasks, update tasks, update task status/progress, and comment on tasks.
- Project `admin` can do everything a member can, plus archive/delete project tasks and manage task assignments.
- Project `owner` can do everything an admin can, plus delete the project and change member roles.
- Standalone and org-level tasks are controlled by task `creator`.
- Task assignee does not grant task access in v1. Assignment remains task metadata for now.
- Team assignment does not grant task access in v1. Team membership is deferred until a team model exists.
- `ListProjects` should return only projects where the caller has `project#can_view`.
- `ListTasksByCreator` should be self-only in v1.
- `ListTasksByOrganization` should wait for org-level OpenFGA integration or be restricted until org permissions exist.

## Draft Model

The draft model lives at [task_model.fga](/home/legaz/relay/packages/core/coreopenfga/models/task_model.fga:1).

Important choices in the draft:

- `project` is the source of truth for project-scoped task access
- standalone and org-level tasks stay creator-managed, because that is what current code does
- `assignee` exists as a relation, but it does not grant access yet
- `task_comment` is modeled explicitly because comment update/delete is author-only

## Check Mapping

The first OpenFGA authorizer pass can keep the current `RequireProjectRole` and `RequireTaskRole` interface, then map those role checks to model checks internally.

| Current check | OpenFGA check |
| --- | --- |
| project `member` | `project:<id>#can_view@user:<id>` |
| project `admin` | `project:<id>#can_manage_tasks@user:<id>` |
| project `owner` | `project:<id>#can_delete@user:<id>` |
| task `member` | `task:<id>#can_view@user:<id>` or action-specific `can_edit` / `can_comment` |
| task `admin` | `task:<id>#can_manage@user:<id>` |

Longer term, service methods should call permission names directly instead of passing role names.

## Tuple Sync Plan

After successful DB writes, `task-service` should sync tuples like this:

- `CreateProject`: write `project:<id>#owner@user:<created_by>`
- `AddProjectMember`: write the matching `owner` / `admin` / `member` tuple
- `UpdateProjectMemberRole`: delete old role tuple, write new role tuple
- `RemoveProjectMember`: delete the role tuple
- `CreateTask` in project: write `task:<id>#creator@user:<created_by>` and `task:<id>#parent_project@project:<project_id>`
- `CreateTask` standalone or org-level: write `task:<id>#creator@user:<created_by>`
- `AssignTask`: write `task:<id>#assignee@user:<id>` for user assignees; team assignee tuples can be written later when team permissions exist
- `UnassignTask`: delete the matching assignee tuple
- `CreateTaskComment`: write `task_comment:<id>#author@user:<author_id>` and `task_comment:<id>#parent_task@task:<task_id>`

## Deferred Work

These items should not block v1, but they should be handled before treating authorization as complete:

- add org-level authorization for organization-scoped projects and tasks
- add team membership relations before team assignments affect access
- decide whether assignees should eventually receive task `can_view` or `can_edit`
- replace role-shaped service calls with permission-shaped service calls after the OpenFGA authorizer is stable
