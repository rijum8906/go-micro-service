CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    
    title varchar(255) NOT NULL,
    description text,
    priority varchar(50),
    due_at timestamptz, 
    
    created_by_user_id uuid NOT NULL,
    assigned_to_user_id uuid,
    team_id uuid,
    status varchar(50) NOT NULL default 'todo',
    
     -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);


-- Index
CREATE INDEX idx_task_status ON tasks(status);
CREATE INDEX idx_task_assigned_to_user_id ON tasks(assigned_to_user_id);
CREATE INDEX idx_task_created_by_user_id ON tasks(created_by_user_id);
CREATE INDEX idx_task_team_id ON tasks(team_id);
CREATE INDEX idx_task_due_at ON tasks(due_at);