-- Create "notifications" table
CREATE TABLE "notifications" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "recepient_email" character varying(255) NOT NULL,
  "recepient_user_id" uuid NOT NULL,
  "message_data" jsonb NOT NULL,
  "status" character varying(50) NOT NULL DEFAULT 'pending',
  "template_type" character varying(255) NOT NULL,
  "retry_count" integer NOT NULL DEFAULT 0,
  "last_error" text NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create index "idx_notifications_recipient" to table: "notifications"
CREATE INDEX "idx_notifications_recipient" ON "notifications" ("recepient_user_id");
-- Create index "idx_notifications_status" to table: "notifications"
CREATE INDEX "idx_notifications_status" ON "notifications" ("status");
