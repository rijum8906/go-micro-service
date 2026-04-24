create extension if not exists pgcrypto;

-- =====================================================
-- HELPER FUNCTIONS
-- =====================================================

create or replace function set_updated_at()
returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

-- =====================================================
-- PROJECTS
-- =====================================================
create table if not exists projects (
    id uuid primary key default gen_random_uuid(),
    
    -- NULL here means this is a personal project, not tied to any org
    organization_id uuid,
    created_by uuid not null,
    
    name varchar(255) not null,
    description text,
    status varchar(30) not null default 'active'
        check (status in ('active', 'archived', 'completed')),
    
    -- When these get set, the project moves to a different state
    archived_at timestamptz,
    completed_at timestamptz,
    deleted_at timestamptz,
    deleted_by uuid,
    
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    
    constraint chk_projects_deleted_by check (
        deleted_at is not null or deleted_by is null
    )
);

comment on table projects is 'Projects are containers for tasks. Think of them as folders or workstreams.';
comment on column projects.organization_id is 'If null, this is a personal project. If set, it belongs to an org.';
comment on column projects.status is 'active = we are working on it, archived = frozen but visible, completed = done and dusted';
comment on column projects.deleted_at is 'Soft delete. Once this is set, project is gone from UI but can be restored.';

-- Most common query: show me all active projects in this org
create index idx_projects_organization_lookup
    on projects (organization_id, status)
    where deleted_at is null and archived_at is null;

-- Needed for "show me all projects I created" dashboard view
create index idx_projects_created_by
    on projects (created_by)
    where deleted_at is null;

create trigger trg_projects_updated_at
    before update on projects
    for each row
    execute function set_updated_at();

-- =====================================================
-- PROJECT MEMBERSHIPS
-- =====================================================
create table if not exists project_memberships (
    id uuid primary key default gen_random_uuid(),
    
    project_id uuid not null references projects(id) on delete cascade,
    user_id uuid not null,
    role varchar(30) not null default 'member'
        check (role in ('owner', 'admin', 'member')),
    
    -- When did they join, and when did they leave (if ever)
    joined_at timestamptz not null default now(),
    left_at timestamptz,
    
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

comment on table project_memberships is 'Who belongs to which project, and what can they do there.';
comment on column project_memberships.role is 'owner = can delete project, admin = can add/remove members, member = can see and do tasks';
comment on column project_memberships.left_at is 'Null means they are still in the project. Once set, they are out.';

-- A user cannot be added twice to the same project while they are still active
create unique index uq_project_memberships_active
    on project_memberships (project_id, user_id)
    where left_at is null;

-- Dashboard query: show me all projects where I am a member
create index idx_project_memberships_user_lookup
    on project_memberships (user_id)
    include (project_id, role)
    where left_at is null;

-- Project settings page: show me all members of this project
create index idx_project_memberships_project_lookup
    on project_memberships (project_id)
    include (user_id, role)
    where left_at is null;

create trigger trg_project_memberships_updated_at
    before update on project_memberships
    for each row
    execute function set_updated_at();

-- =====================================================
-- TASKS
-- =====================================================
create table if not exists tasks (
    id uuid primary key default gen_random_uuid(),
    
    -- A task can live in three places:
    -- 1. Personal: both null
    -- 2. Org-level: organization_id set, project_id null
    -- 3. Inside a project: project_id set and organization_id left null
    organization_id uuid,
    project_id uuid references projects(id) on delete restrict,
    parent_task_id uuid references tasks(id) on delete set null,
    
    -- Who made this task, and who last touched it
    created_by uuid not null,
    updated_by uuid,
    
    title varchar(255) not null,
    description text,
    status varchar(30) not null default 'pending'
        check (status in ('pending', 'in_progress', 'blocked', 'completed', 'cancelled')),
    priority varchar(20) not null default 'medium'
        check (priority in ('low', 'medium', 'high', 'urgent')),
    progress_percent smallint not null default 0
        check (progress_percent >= 0 and progress_percent <= 100),
    
    -- Timeline tracking
    started_at timestamptz,     -- When someone first moved it to in_progress
    due_at timestamptz,         -- Deadline
    completed_at timestamptz,   -- When status changed to completed
    
    -- Soft delete fields
    archived_at timestamptz,    -- Hidden but not deleted
    deleted_at timestamptz,     -- Actually deleted, but row stays for audit
    deleted_by uuid,
    
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    
    -- Project tasks derive organization scope from the project row.
    constraint chk_task_scope check (
        project_id is null or organization_id is null
    ),
    constraint chk_task_completion check (
        (status = 'completed' and completed_at is not null) or
        (status <> 'completed' and completed_at is null)
    ),
    constraint chk_task_deleted_by check (
        deleted_at is not null or deleted_by is null
    )
);

comment on table tasks is 'The actual work. Tasks can be broken down into subtasks via parent_task_id.';
comment on column tasks.parent_task_id is 'Points to another task. Null means this is a top-level task.';
comment on column tasks.progress_percent is 'Manual 0-100 slider. Helps with tracking without forcing status changes.';
comment on column tasks.deleted_at is 'Soft delete. Task disappears from UI but we keep the record for history.';

-- Project view: show me all tasks in this project, grouped by status
create index idx_tasks_project_lookup
    on tasks (project_id, status)
    where deleted_at is null;

-- Org-level view: show me all tasks that belong directly to the org (not inside any project)
create index idx_tasks_org_lookup
    on tasks (organization_id, status)
    where deleted_at is null and project_id is null;

-- "My tasks" view: show me all tasks I created
create index idx_tasks_creator_lookup
    on tasks (created_by, status)
    where deleted_at is null;

-- When viewing a parent task, we need to quickly fetch all its subtasks
create index idx_tasks_parent_lookup
    on tasks (parent_task_id)
    where deleted_at is null;

-- Dashboard alert: show me all overdue tasks that are not yet completed
create index idx_tasks_due_lookup
    on tasks (due_at)
    where deleted_at is null and completed_at is null;

-- Sorting by priority is common in backlog views
create index idx_tasks_priority_status
    on tasks (priority, status)
    where deleted_at is null;

create trigger trg_tasks_updated_at
    before update on tasks
    for each row
    execute function set_updated_at();

-- =====================================================
-- TASK ASSIGNMENTS
-- =====================================================
create table if not exists task_assignments (
    id uuid primary key default gen_random_uuid(),
    
    task_id uuid not null references tasks(id) on delete cascade,
    
    -- Two types: 'user' means one person, 'team' means anyone from that team can claim it
    assignee_type varchar(20) not null
        check (assignee_type in ('user', 'team')),
    assignee_id uuid not null,  -- Points to users table or teams table depending on type
    assigned_by uuid not null,   -- Who made this assignment
    
    assigned_at timestamptz not null default now(),
    unassigned_at timestamptz,   -- When this assignment ended. Null means it's still active.
    
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

comment on table task_assignments is 'Who is responsible for this task. Can be a single person or a whole team.';
comment on column task_assignments.assignee_type is 'user = assigned to a specific person, team = assigned to a team (anyone can pick up)';
comment on column task_assignments.unassigned_at is 'When this assignment was removed. We keep history to know who worked on what and when.';

-- A task cannot be assigned to the same user or team twice while the assignment is active
create unique index uq_task_assignments_active
    on task_assignments (task_id, assignee_type, assignee_id)
    where unassigned_at is null;

-- Dashboard: find all tasks assigned to me (or my team)
create index idx_task_assignments_assignee_lookup
    on task_assignments (assignee_type, assignee_id)
    include (task_id, assigned_by, assigned_at)
    where unassigned_at is null;

-- Task detail page: show me all current and past assignments for this task
create index idx_task_assignments_task_lookup
    on task_assignments (task_id, assignee_type)
    include (assignee_id, assigned_by, assigned_at)
    where unassigned_at is null;

-- Audit: who was responsible when this task failed
create index idx_task_assignments_history
    on task_assignments (task_id, assigned_at desc)
    where unassigned_at is not null;

create trigger trg_task_assignments_updated_at
    before update on task_assignments
    for each row
    execute function set_updated_at();

-- =====================================================
-- TASK COMMENTS
-- =====================================================
create table if not exists task_comments (
    id uuid primary key default gen_random_uuid(),
    
    task_id uuid not null references tasks(id) on delete cascade,
    author_id uuid not null,
    body text not null,
    
    -- Edit tracking
    edited_at timestamptz,
    edit_count integer not null default 0
        check (edit_count >= 0),
    
    -- Soft delete: comment is hidden but we keep the record for moderation/audit
    deleted_at timestamptz,
    deleted_by uuid,
    
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    
    constraint chk_task_comments_deleted_by check (
        deleted_at is not null or deleted_by is null
    )
);

comment on table task_comments is 'Discussion on a task. Similar to GitHub issue comments.';
comment on column task_comments.edit_count is 'How many times the user edited this comment. Helps with moderation.';
comment on column task_comments.deleted_at is 'Soft delete. Comment disappears from UI but we keep it for legal/audit reasons.';

-- Task detail page: show all comments, oldest first
create index idx_task_comments_task_lookup
    on task_comments (task_id, created_at asc)
    where deleted_at is null;

-- User profile: show recent comments from this person across all tasks
create index idx_task_comments_author_lookup
    on task_comments (author_id, created_at desc)
    where deleted_at is null;

create trigger trg_task_comments_updated_at
    before update on task_comments
    for each row
    execute function set_updated_at();
