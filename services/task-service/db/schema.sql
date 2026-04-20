create table if not exists tasks (
    id uuid primary key default gen_random_uuid(),

    assigned_by uuid not null,

    title varchar(255) not null,
    description text,
    status varchar(50) not null default 'pending',
    -- Timestamps
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

