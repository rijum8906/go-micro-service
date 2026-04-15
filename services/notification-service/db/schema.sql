CREATE TABLE notification_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    job_subject varchar(128) NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'received',

    raw_payload text NOT NULL,
    error_message text,

    received_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT notification_jobs_status_check
        CHECK (status IN ('received', 'processing', 'completed', 'failed', 'invalid_payload'))
);

CREATE TABLE notification_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    job_id uuid REFERENCES notification_jobs(id) ON DELETE SET NULL,

    channel varchar(32) NOT NULL DEFAULT 'email',
    notification_type varchar(64) NOT NULL,
    template_type varchar(128),

    recipient_email varchar(320),
    recipient_name varchar(255),
    subject varchar(255),

    status varchar(32) NOT NULL DEFAULT 'pending',
    last_error text,

    provider varchar(64),
    provider_message_id varchar(255),

    queued_at timestamptz NOT NULL DEFAULT now(),
    sent_at timestamptz,
    failed_at timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT notification_deliveries_channel_check
        CHECK (channel IN ('email', 'sms', 'push', 'webhook')),

    CONSTRAINT notification_deliveries_status_check
        CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'cancelled'))
);

CREATE TABLE notification_delivery_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    delivery_id uuid NOT NULL REFERENCES notification_deliveries(id) ON DELETE CASCADE,

    attempt_no int NOT NULL,
    status varchar(32) NOT NULL,

    error_message text,
    error_details jsonb NOT NULL DEFAULT '{}'::jsonb,

    provider varchar(64),
    provider_message_id varchar(255),

    started_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT notification_delivery_attempts_status_check
        CHECK (status IN ('success', 'failed')),

    CONSTRAINT notification_delivery_attempts_attempt_no_check
        CHECK (attempt_no > 0),

    CONSTRAINT notification_delivery_attempts_unique_attempt
        UNIQUE (delivery_id, attempt_no)
);

CREATE INDEX idx_notification_jobs_subject ON notification_jobs(job_subject);
CREATE INDEX idx_notification_jobs_status ON notification_jobs(status);
CREATE INDEX idx_notification_deliveries_status ON notification_deliveries(status);
CREATE INDEX idx_notification_deliveries_recipient_email ON notification_deliveries(recipient_email);
CREATE INDEX idx_notification_delivery_attempts_delivery_id ON notification_delivery_attempts(delivery_id);
