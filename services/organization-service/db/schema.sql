CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- Helper Functions
-- =====================================================

-- Automatically update updated_at timestamp on row changes
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Organizations (tenants/workspaces)
-- =====================================================
CREATE TABLE organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    status varchar(30) NOT NULL DEFAULT 'active',
    slug varchar(80) UNIQUE NOT NULL,
    description text,
    logo_url text,
    created_by uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Ensures slugs are consistently lowercase to prevent case-sensitive duplicates
    CONSTRAINT chk_organizations_slug_lowercase CHECK (slug = lower(slug)),

    -- Slugs should only contain lowercase letters, numbers, and hyphens
    CONSTRAINT chk_organizations_slug_format CHECK (slug ~ '^[a-z0-9-]+$')
);

COMMENT ON TABLE organizations IS 'Core tenant/workspace entity that contains teams and members';
COMMENT ON COLUMN organizations.slug IS 'URL-friendly identifier, unique across all organizations';
COMMENT ON COLUMN organizations.created_by IS 'User ID who created this organization';

CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================
-- Organization Memberships (users belonging to orgs)
-- =====================================================
CREATE TABLE organization_memberships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    role varchar(30) NOT NULL DEFAULT 'member'
        CHECK (role IN ('owner', 'admin', 'member')),
    status varchar(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'left')),
    invited_by uuid,
    joined_at timestamptz NOT NULL DEFAULT now(),
    left_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Prevents duplicate memberships for the same user in an organization
    CONSTRAINT uq_organization_memberships_org_user UNIQUE (organization_id, user_id)
);

COMMENT ON TABLE organization_memberships IS 'Junction table linking users to organizations with role and status';
COMMENT ON COLUMN organization_memberships.role IS 'owner=full org control, admin=manage members/teams, member=standard access';
COMMENT ON COLUMN organization_memberships.status IS 'active=current member, suspended=temporarily blocked, left=voluntarily departed';

-- For finding all orgs a user belongs to
CREATE INDEX idx_organization_memberships_user_id
    ON organization_memberships (user_id);

-- For finding all members of an org (most common query)
CREATE INDEX idx_organization_memberships_organization_id
    ON organization_memberships (organization_id);

CREATE TRIGGER trg_organization_memberships_updated_at
    BEFORE UPDATE ON organization_memberships
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================
-- Teams (sub-groups within organizations)
-- =====================================================
CREATE TABLE organization_teams (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    created_by uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    deleted_by uuid,

    -- Team names must be unique within an organization (can't have two "Engineering" teams)
    CONSTRAINT uq_organization_teams_org_name UNIQUE (organization_id, name)
);

COMMENT ON TABLE organization_teams IS 'Teams are sub-groups within an organization for fine-grained access control';
COMMENT ON COLUMN organization_teams.name IS 'Team name, must be unique per organization (e.g., "Engineering", "Sales")';


-- Essential for "show all teams in this organization" queries
CREATE INDEX idx_organization_teams_organization_id
    ON organization_teams (organization_id);

CREATE TRIGGER trg_organization_teams_updated_at
    BEFORE UPDATE ON organization_teams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================
-- Team Memberships (which org members belong to which teams)
-- =====================================================
CREATE TABLE organization_team_memberships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id uuid NOT NULL,
    membership_id uuid NOT NULL,  -- References organization_memberships, not users directly
    role varchar(30) NOT NULL DEFAULT 'member',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,  -- Soft delete: when a member leaves a team
    deleted_by uuid REFERENCES organization_memberships(id),

    CONSTRAINT fk_organization_team_memberships_team
        FOREIGN KEY (team_id)
        REFERENCES organization_teams(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_organization_team_memberships_membership
        FOREIGN KEY (membership_id)
        REFERENCES organization_memberships(id)
        ON DELETE CASCADE
);

COMMENT ON TABLE organization_team_memberships IS 'Links organization members to teams. Uses membership_id (not user_id) to respect org-level roles.';
COMMENT ON COLUMN organization_team_memberships.membership_id IS 'References organization_memberships.id - a user must be an org member before joining a team';
COMMENT ON COLUMN organization_team_memberships.deleted_at IS 'Soft delete timestamp. When non-NULL, the member is no longer in the team.';

-- Prevents duplicate active team memberships (a user can't be in the same team twice)
-- The WHERE clause makes this a partial unique index - only enforces for active records
CREATE UNIQUE INDEX uq_organization_team_memberships_active
    ON organization_team_memberships (organization_id, team_id, membership_id)
    WHERE deleted_at IS NULL;

-- Fast lookup: "Find all teams that a user belongs to in an organization"
-- INCLUDE stores team_id and role directly in the index, avoiding a table lookup
CREATE INDEX idx_organization_team_memberships_user_lookup
    ON organization_team_memberships (organization_id, membership_id)
    INCLUDE (team_id, role)
    WHERE deleted_at IS NULL;

-- Fast lookup: "Find all members of a specific team"
-- INCLUDE stores membership_id and role, making the query fully index-covered
CREATE INDEX idx_organization_team_memberships_team_lookup
    ON organization_team_memberships (organization_id, team_id)
    INCLUDE (membership_id, role)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_organization_team_memberships_updated_at
    BEFORE UPDATE ON organization_team_memberships
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================
-- Organization Invitations (pending invites to join orgs)
-- =====================================================
CREATE TABLE organization_invitations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email varchar(320) NOT NULL,
    role varchar(30) NOT NULL DEFAULT 'member'
        CHECK (role IN ('owner', 'admin', 'member')),
    status varchar(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),
    invited_by uuid NOT NULL,
    token_hash varchar(255) UNIQUE NOT NULL,  -- Store hash, not raw token for security
    expires_at timestamptz NOT NULL,           -- Tokens expire after N days
    accepted_by uuid,
    accepted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE organization_invitations IS 'Tracks pending invitations to join an organization';
COMMENT ON COLUMN organization_invitations.token_hash IS 'SHA256 hash of the invitation token. Never store raw tokens.';
COMMENT ON COLUMN organization_invitations.expires_at IS 'Invitations are invalid after this timestamp (typically 7 days)';

-- Find all pending invites for an organization (admin UI)
CREATE INDEX idx_organization_invitations_organization_id
    ON organization_invitations (organization_id);

-- Find all pending invites for a specific email (during signup/accept flow)
CREATE INDEX idx_organization_invitations_email
    ON organization_invitations (email);

-- =====================================================
-- Organization Audit Logs (compliance & debugging)
-- =====================================================
CREATE TABLE organization_audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id uuid,  -- NULL for system actions (e.g., automated cleanup)
    action varchar(100) NOT NULL,
    target_type varchar(50),
    target_id uuid,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE organization_audit_logs IS 'Immutable audit trail for security, compliance, and debugging';
COMMENT ON COLUMN organization_audit_logs.action IS 'Verb like "member_added", "role_changed", "settings_updated"';
COMMENT ON COLUMN organization_audit_logs.target_type IS 'Entity type: "user", "team", "membership", "organization"';
COMMENT ON COLUMN organization_audit_logs.metadata IS 'Additional context like {old_role: "member", new_role: "admin"}';

-- Most common query pattern: "Show recent activity for an organization"
-- This composite index covers both org filtering and time ordering
CREATE INDEX idx_organization_audit_logs_org_time
    ON organization_audit_logs (organization_id, created_at DESC);

-- We drop the individual indexes since this composite index serves both purposes
DROP INDEX IF EXISTS idx_organization_audit_logs_organization_id;
DROP INDEX IF EXISTS idx_organization_audit_logs_created_at;

-- =====================================================
-- Optional: Additional Useful Indexes
-- =====================================================

-- For "find all actions performed by a specific user" (investigations)
CREATE INDEX idx_organization_audit_logs_actor
    ON organization_audit_logs (actor_user_id, created_at DESC)
    WHERE actor_user_id IS NOT NULL;

-- For "show all changes to a specific entity" (e.g., document history)
CREATE INDEX idx_organization_audit_logs_target
    ON organization_audit_logs (target_type, target_id, created_at DESC)
    WHERE target_id IS NOT NULL;

-- For JSONB queries inside metadata (if you frequently search inside it)
-- CREATE INDEX idx_organization_audit_logs_metadata ON organization_audit_logs USING gin (metadata);
