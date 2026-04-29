create table if not exists notifications (
    id uuid primary key default gen_random_uuid(),

    recepient_email varchar(255) not null,
    recepient_user_id uuid not null,
    message_data jsonb not null,

    status varchar(50) not null default 'pending',
    template_type varchar(255) not null,
    -- Retry tracking
    retry_count int not null default 0,
    last_error text,
    -- Timestamps
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- Indexing for your worker queries
create index idx_notifications_status on notifications(status);
create index idx_notifications_recipient on notifications(recepient_user_id);
