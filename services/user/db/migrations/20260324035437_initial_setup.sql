-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" character varying(255) NOT NULL,
  "password_hash" character varying(255) NULL,
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "two_factor_enabled" boolean NOT NULL DEFAULT false,
  "two_factor_enabled_at" timestamptz NULL,
  "email_verified_at" timestamptz NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);
-- Create "oauths" table
CREATE TABLE "oauths" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "provider" character varying(255) NOT NULL,
  "provider_id" character varying(255) NOT NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "oauths_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "profiles" table
CREATE TABLE "profiles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "first_name" character varying(255) NOT NULL,
  "last_name" character varying(255) NOT NULL,
  "avatar_url" text NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "profiles_user_id_key" UNIQUE ("user_id"),
  CONSTRAINT "profiles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "refresh_token_hash" character varying(255) NOT NULL,
  "user_agent" character varying(255) NOT NULL,
  "ip_addr" character varying(255) NOT NULL,
  "device_id" character varying(255) NOT NULL,
  "last_login_at" timestamptz NULL DEFAULT now(),
  "expires_at" timestamptz NOT NULL,
  "is_revoked" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_sessions_active" to table: "sessions"
CREATE INDEX "idx_sessions_active" ON "sessions" ("user_id") WHERE (is_revoked = false);
-- Create index "idx_sessions_expires" to table: "sessions"
CREATE INDEX "idx_sessions_expires" ON "sessions" ("expires_at");
-- Create index "idx_sessions_last_login" to table: "sessions"
CREATE INDEX "idx_sessions_last_login" ON "sessions" ("last_login_at" DESC);
-- Create index "idx_sessions_refresh_token" to table: "sessions"
CREATE UNIQUE INDEX "idx_sessions_refresh_token" ON "sessions" ("refresh_token_hash");
-- Create index "idx_sessions_user_id" to table: "sessions"
CREATE INDEX "idx_sessions_user_id" ON "sessions" ("user_id");
