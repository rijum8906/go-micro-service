-- Create "set_updated_at" function
CREATE FUNCTION "set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;
-- Create "organizations" table
CREATE TABLE "organizations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NOT NULL,
  "status" character varying(30) NOT NULL DEFAULT 'active',
  "slug" character varying(80) NOT NULL,
  "description" text NULL,
  "logo_url" text NULL,
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid,
  "archived_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "organizations_slug_key" UNIQUE ("slug"),
  CONSTRAINT "chk_organizations_slug_format" CHECK ((slug)::text ~ '^[a-z0-9-]+$'::text),
  CONSTRAINT "chk_organizations_slug_lowercase" CHECK ((slug)::text = lower((slug)::text)),
  CONSTRAINT "organizations_status_check" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'archived'::character varying, 'deleted'::character varying])::text[]))
);
-- Set comment to table: "organizations"
COMMENT ON TABLE "organizations" IS 'Core tenant/workspace entity that contains teams and members';
-- Set comment to column: "slug" on table: "organizations"
COMMENT ON COLUMN "organizations"."slug" IS 'URL-friendly identifier, unique across all organizations';
-- Set comment to column: "created_by" on table: "organizations"
COMMENT ON COLUMN "organizations"."created_by" IS 'User ID who created this organization';
-- Create "organization_memberships" table
CREATE TABLE "organization_memberships" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "organization_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" character varying(30) NOT NULL DEFAULT 'member',
  "status" character varying(30) NOT NULL DEFAULT 'active',
  "joined_at" timestamptz NOT NULL DEFAULT now(),
  "left_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_organization_memberships_org_user" UNIQUE ("organization_id", "user_id"),
  CONSTRAINT "organization_memberships_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "organization_memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "organization_memberships_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "organization_memberships_status_check" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'suspended'::character varying, 'left'::character varying])::text[]))
);
-- Create index "idx_organization_memberships_organization_id" to table: "organization_memberships"
CREATE INDEX "idx_organization_memberships_organization_id" ON "organization_memberships" ("organization_id");
-- Create index "idx_organization_memberships_user_id" to table: "organization_memberships"
CREATE INDEX "idx_organization_memberships_user_id" ON "organization_memberships" ("user_id");
-- Set comment to table: "organization_memberships"
COMMENT ON TABLE "organization_memberships" IS 'Junction table linking users to organizations with role and status';
-- Set comment to column: "role" on table: "organization_memberships"
COMMENT ON COLUMN "organization_memberships"."role" IS 'owner=full org control, admin=manage members/teams, member=standard access';
-- Set comment to column: "status" on table: "organization_memberships"
COMMENT ON COLUMN "organization_memberships"."status" IS 'active=current member, suspended=temporarily blocked, left=voluntarily departed';
-- Create trigger "trg_organization_memberships_updated_at"
CREATE TRIGGER "trg_organization_memberships_updated_at" BEFORE UPDATE ON "organization_memberships" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "organization_teams" table
CREATE TABLE "organization_teams" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "organization_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_organization_teams_org_name" UNIQUE ("organization_id", "name"),
  CONSTRAINT "organization_teams_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "organization_memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "organization_teams_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_organization_teams_organization_id" to table: "organization_teams"
CREATE INDEX "idx_organization_teams_organization_id" ON "organization_teams" ("organization_id");
-- Set comment to table: "organization_teams"
COMMENT ON TABLE "organization_teams" IS 'Teams are sub-groups within an organization for fine-grained access control';
-- Set comment to column: "name" on table: "organization_teams"
COMMENT ON COLUMN "organization_teams"."name" IS 'Team name, must be unique per organization (e.g., "Engineering", "Sales")';
-- Create "organization_team_memberships" table
CREATE TABLE "organization_team_memberships" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "organization_id" uuid NOT NULL,
  "team_id" uuid NOT NULL,
  "membership_id" uuid NOT NULL,
  "role" character varying(30) NOT NULL DEFAULT 'member',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_organization_team_memberships_membership" FOREIGN KEY ("membership_id") REFERENCES "organization_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_organization_team_memberships_team" FOREIGN KEY ("team_id") REFERENCES "organization_teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "organization_team_memberships_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "organization_team_memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "organization_team_memberships_membership_id_fkey" FOREIGN KEY ("membership_id") REFERENCES "organization_memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "organization_team_memberships_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "organization_team_memberships_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "organization_teams" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_organization_team_memberships_team_lookup" to table: "organization_team_memberships"
CREATE INDEX "idx_organization_team_memberships_team_lookup" ON "organization_team_memberships" ("organization_id", "team_id") INCLUDE ("membership_id", "role") WHERE (deleted_at IS NULL);
-- Create index "idx_organization_team_memberships_user_lookup" to table: "organization_team_memberships"
CREATE INDEX "idx_organization_team_memberships_user_lookup" ON "organization_team_memberships" ("organization_id", "membership_id") INCLUDE ("team_id", "role") WHERE (deleted_at IS NULL);
-- Create index "uq_organization_team_memberships_active" to table: "organization_team_memberships"
CREATE UNIQUE INDEX "uq_organization_team_memberships_active" ON "organization_team_memberships" ("organization_id", "team_id", "membership_id") WHERE (deleted_at IS NULL);
-- Set comment to table: "organization_team_memberships"
COMMENT ON TABLE "organization_team_memberships" IS 'Links organization members to teams. Uses membership_id (not user_id) to respect org-level roles.';
-- Set comment to column: "membership_id" on table: "organization_team_memberships"
COMMENT ON COLUMN "organization_team_memberships"."membership_id" IS 'References organization_memberships.id - a user must be an org member before joining a team';
-- Set comment to column: "deleted_at" on table: "organization_team_memberships"
COMMENT ON COLUMN "organization_team_memberships"."deleted_at" IS 'Soft delete timestamp. When non-NULL, the member is no longer in the team.';
-- Create trigger "trg_organization_team_memberships_updated_at"
CREATE TRIGGER "trg_organization_team_memberships_updated_at" BEFORE UPDATE ON "organization_team_memberships" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create trigger "trg_organization_teams_updated_at"
CREATE TRIGGER "trg_organization_teams_updated_at" BEFORE UPDATE ON "organization_teams" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create trigger "trg_organizations_updated_at"
CREATE TRIGGER "trg_organizations_updated_at" BEFORE UPDATE ON "organizations" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "organization_audit_logs" table
CREATE TABLE "organization_audit_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "organization_id" uuid NOT NULL,
  "actor_user_id" uuid NULL,
  "action" character varying(100) NOT NULL,
  "target_type" character varying(50) NULL,
  "target_id" uuid NULL,
  "metadata" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "organization_audit_logs_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_organization_audit_logs_actor" to table: "organization_audit_logs"
CREATE INDEX "idx_organization_audit_logs_actor" ON "organization_audit_logs" ("actor_user_id", "created_at" DESC) WHERE (actor_user_id IS NOT NULL);
-- Create index "idx_organization_audit_logs_org_time" to table: "organization_audit_logs"
CREATE INDEX "idx_organization_audit_logs_org_time" ON "organization_audit_logs" ("organization_id", "created_at" DESC);
-- Create index "idx_organization_audit_logs_target" to table: "organization_audit_logs"
CREATE INDEX "idx_organization_audit_logs_target" ON "organization_audit_logs" ("target_type", "target_id", "created_at" DESC) WHERE (target_id IS NOT NULL);
-- Set comment to table: "organization_audit_logs"
COMMENT ON TABLE "organization_audit_logs" IS 'Immutable audit trail for security, compliance, and debugging';
-- Set comment to column: "action" on table: "organization_audit_logs"
COMMENT ON COLUMN "organization_audit_logs"."action" IS 'Verb like "member_added", "role_changed", "settings_updated"';
-- Set comment to column: "target_type" on table: "organization_audit_logs"
COMMENT ON COLUMN "organization_audit_logs"."target_type" IS 'Entity type: "user", "team", "membership", "organization"';
-- Set comment to column: "metadata" on table: "organization_audit_logs"
COMMENT ON COLUMN "organization_audit_logs"."metadata" IS 'Additional context like {old_role: "member", new_role: "admin"}';
-- Create "organization_invitations" table
CREATE TABLE "organization_invitations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "organization_id" uuid NOT NULL,
  "email" character varying(320) NOT NULL,
  "role" character varying(30) NOT NULL DEFAULT 'member',
  "status" character varying(30) NOT NULL DEFAULT 'pending',
  "invited_by" uuid NOT NULL,
  "token_hash" character varying(255) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "accepted_by" uuid NULL,
  "accepted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "organization_invitations_token_hash_key" UNIQUE ("token_hash"),
  CONSTRAINT "organization_invitations_invited_by_fkey" FOREIGN KEY ("invited_by") REFERENCES "organization_memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "organization_invitations_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "organization_invitations_status_check" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'accepted'::character varying, 'revoked'::character varying, 'expired'::character varying])::text[]))
);
-- Create index "idx_organization_invitations_email" to table: "organization_invitations"
CREATE INDEX "idx_organization_invitations_email" ON "organization_invitations" ("email");
-- Create index "idx_organization_invitations_organization_id" to table: "organization_invitations"
CREATE INDEX "idx_organization_invitations_organization_id" ON "organization_invitations" ("organization_id");
-- Set comment to table: "organization_invitations"
COMMENT ON TABLE "organization_invitations" IS 'Tracks pending invitations to join an organization';
-- Set comment to column: "token_hash" on table: "organization_invitations"
COMMENT ON COLUMN "organization_invitations"."token_hash" IS 'SHA256 hash of the invitation token. Never store raw tokens.';
-- Set comment to column: "expires_at" on table: "organization_invitations"
COMMENT ON COLUMN "organization_invitations"."expires_at" IS 'Invitations are invalid after this timestamp (typically 7 days)';
