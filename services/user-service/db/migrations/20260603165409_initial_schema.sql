-- Create "set_updated_at" function
CREATE FUNCTION "set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" character varying(255) NOT NULL,
  "status" character varying(30) NOT NULL DEFAULT 'active',
  "password_hash" character varying(255) NULL,
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "email_verified_at" timestamptz NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email"),
  CONSTRAINT "users_status_check" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'deleted'::character varying])::text[]))
);
-- Create index "idx_users_email" to table: "users"
CREATE INDEX "idx_users_email" ON "users" ("email");
-- Create index "idx_users_is_email_verified" to table: "users"
CREATE INDEX "idx_users_is_email_verified" ON "users" ("is_email_verified");
-- Create index "idx_users_status" to table: "users"
CREATE INDEX "idx_users_status" ON "users" ("status");
-- Set comment to column: "status" on table: "users"
COMMENT ON COLUMN "users"."status" IS 'The status of the user (active, inactive, deleted)';
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
-- Create index "idx_oauths_provider_provider_id" to table: "oauths"
CREATE INDEX "idx_oauths_provider_provider_id" ON "oauths" ("provider", "provider_id");
-- Create index "idx_oauths_user_id" to table: "oauths"
CREATE INDEX "idx_oauths_user_id" ON "oauths" ("user_id");
-- Create trigger "set_updated_at_oauths"
CREATE TRIGGER "set_updated_at_oauths" BEFORE UPDATE ON "oauths" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
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
-- Create index "idx_profiles_user_id" to table: "profiles"
CREATE INDEX "idx_profiles_user_id" ON "profiles" ("user_id");
-- Create trigger "set_updated_at_profiles"
CREATE TRIGGER "set_updated_at_profiles" BEFORE UPDATE ON "profiles" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
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
-- Create index "idx_sessions_expires" to table: "sessions"
CREATE INDEX "idx_sessions_expires" ON "sessions" ("expires_at");
-- Create index "idx_sessions_is_revoked" to table: "sessions"
CREATE INDEX "idx_sessions_is_revoked" ON "sessions" ("is_revoked");
-- Create index "idx_sessions_last_login" to table: "sessions"
CREATE INDEX "idx_sessions_last_login" ON "sessions" ("last_login_at" DESC);
-- Create index "idx_sessions_refresh_token" to table: "sessions"
CREATE INDEX "idx_sessions_refresh_token" ON "sessions" ("refresh_token_hash");
-- Create index "idx_sessions_user_id" to table: "sessions"
CREATE INDEX "idx_sessions_user_id" ON "sessions" ("user_id");
-- Create trigger "set_updated_at_sessions"
CREATE TRIGGER "set_updated_at_sessions" BEFORE UPDATE ON "sessions" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "two_factors" table
CREATE TABLE "two_factors" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "method" character varying(255) NOT NULL,
  "secret" character varying(255) NOT NULL,
  "is_enabled" boolean NULL DEFAULT false,
  "is_primary" boolean NULL DEFAULT true,
  "last_used_at" timestamptz NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "unique_method_per_user" UNIQUE ("user_id", "method"),
  CONSTRAINT "two_factors_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "two_factors_method_check" CHECK ((method)::text = ANY ((ARRAY['totp'::character varying, 'email'::character varying, 'webauthn'::character varying])::text[]))
);
-- Create index "idx_one_primary_per_user" to table: "two_factors"
CREATE UNIQUE INDEX "idx_one_primary_per_user" ON "two_factors" ("user_id") WHERE (is_primary = true);
-- Create index "idx_two_factors_user_id" to table: "two_factors"
CREATE INDEX "idx_two_factors_user_id" ON "two_factors" ("user_id");
-- Create trigger "set_updated_at_two_factors"
CREATE TRIGGER "set_updated_at_two_factors" BEFORE UPDATE ON "two_factors" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create trigger "set_updated_at_users"
CREATE TRIGGER "set_updated_at_users" BEFORE UPDATE ON "users" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
