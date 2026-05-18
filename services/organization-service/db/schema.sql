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
    slug varchar(80) UNIQUE NOT NULL,
    description text,
    logo_url text,
    status varchar(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'archived', 'deleted')),
    created_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    archived_at timestamptz,
    deleted_at timestamptz,

    -- Constraints
    CONSTRAINT chk_organizations_slug_lowercase CHECK (slug = lower(slug)),
    CONSTRAINT chk_organizations_slug_format CHECK (slug ~ '^[a-z0-9-]+$')
);

COMMENT ON TABLE organizations IS 'Core tenant/workspace entity that contains teams and members';
COMMENT ON COLUMN organizations.slug IS 'URL-friendly identifier, unique across all organizations';
COMMENT ON COLUMN organizations.created_by_user_id IS 'User ID who created this organization (references external users table)';
COMMENT ON COLUMN organizations.status IS 'active=normal operation, archived=read-only/hidden, deleted=soft-deleted';

CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Use case: Resolve an organization quickly by its URL slug during routing and lookup.
CREATE INDEX idx_organizations_slug ON organizations (slug);
-- Use case: Filter organizations by lifecycle state for admin lists and cleanup jobs.
CREATE INDEX idx_organizations_status ON organizations (status);
-- Use case: Show recently created organizations without a full table sort.
CREATE INDEX idx_organizations_created_at ON organizations (created_at DESC);

-- =====================================================
-- Organization Memberships (users belonging to orgs)
-- =====================================================
CREATE TABLE organization_memberships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    role varchar(30) NOT NULL DEFAULT 'member',
    status varchar(30) not null default 'active'
        check (status in ('active', 'suspended', 'banned', 'left', 'removed')),
    joined_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Constraints
    CONSTRAINT uq_organization_memberships_org_user UNIQUE (organization_id, user_id)
);

COMMENT ON TABLE organization_memberships IS 'Junction table linking users to organizations with role and status';
COMMENT ON COLUMN organization_memberships.role IS 'owner=full org control, admin=manage members/teams, member=standard access';
COMMENT ON COLUMN organization_memberships.status IS 'active=current member, suspended=temporarily blocked, banned=permanently blocked, left=voluntary departure, removed=admin removal';

-- Use case: List all organizations for a user and check a user's membership history.
CREATE INDEX idx_memberships_user_id ON organization_memberships (user_id);
-- Use case: List and paginate all members in an organization.
CREATE INDEX idx_memberships_organization_id ON organization_memberships (organization_id);
-- Use case: Filter memberships by status for moderation and cleanup views.
CREATE INDEX idx_memberships_status ON organization_memberships (status);
-- Use case: Filter memberships by role for owner/admin/member management views.
CREATE INDEX idx_memberships_role ON organization_memberships (role);

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
    status varchar(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'archived', 'deleted')),
    created_by_mem_id uuid NOT NULL REFERENCES organization_memberships(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    archived_at timestamptz,
    deleted_at timestamptz
);

-- Use case: Enforce one active or archived team name per organization while allowing deleted-name reuse.
CREATE UNIQUE INDEX uq_organization_teams_org_name_active
    ON organization_teams (organization_id, name)
    WHERE (status IN ('active', 'archived'));

COMMENT ON TABLE organization_teams IS 'Teams are sub-groups within an organization for fine-grained access control';
COMMENT ON COLUMN organization_teams.name IS 'Team name, must be unique per organization (excluding deleted teams)';
COMMENT ON COLUMN organization_teams.status IS 'active=normal, archived=hidden/read-only, deleted=soft-deleted';

-- Use case: List all teams that belong to an organization.
CREATE INDEX idx_teams_organization_id ON organization_teams (organization_id);
-- Use case: Filter teams by lifecycle state for active, archived, and deleted views.
CREATE INDEX idx_teams_status ON organization_teams (status);
-- Use case: Find teams created by a specific organization member for audit/admin screens.
CREATE INDEX idx_teams_created_by ON organization_teams (created_by_mem_id);

CREATE TRIGGER trg_organization_teams_updated_at
    BEFORE UPDATE ON organization_teams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =====================================================
-- Team Memberships (which org members belong to which teams)
-- =====================================================
CREATE TABLE organization_team_memberships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id uuid NOT NULL REFERENCES organization_teams(id) ON DELETE CASCADE,
    membership_id uuid NOT NULL REFERENCES organization_memberships(id) ON DELETE CASCADE,
    role varchar(30) NOT NULL DEFAULT 'member',
    status varchar(30) not null default 'active'
        check (status in ('active', 'suspended', 'banned', 'left', 'removed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz -- Added missing column referenced below
);

-- Use case: Prevent duplicate active or suspended team memberships for the same org member.
CREATE UNIQUE INDEX uq_team_memberships_active
    ON organization_team_memberships (team_id, membership_id)
    WHERE (status IN ('active', 'suspended'));

COMMENT ON TABLE organization_team_memberships IS 'Links organization members to teams. Uses membership_id (not user_id) to respect org-level roles.';
COMMENT ON COLUMN organization_team_memberships.membership_id IS 'References organization_memberships.id - a user must be an org member before joining a team';
COMMENT ON COLUMN organization_team_memberships.deleted_at IS 'Soft delete timestamp. When non-NULL, the member is no longer in the team. Audit log contains who/when/why.';

-- Use case: Find all active/suspended teams that an organization member belongs to.
CREATE INDEX idx_team_memberships_user_lookup
    ON organization_team_memberships (organization_id, membership_id)
    INCLUDE (team_id, role)
    WHERE status IN ('active', 'suspended');

-- Use case: Find all active/suspended members of a specific team.
CREATE INDEX idx_team_memberships_team_lookup
    ON organization_team_memberships (team_id)
    INCLUDE (membership_id, role)
    WHERE status IN ('active', 'suspended');

-- Use case: Find team memberships with a deletion timestamp for cleanup and audit review.
CREATE INDEX idx_team_memberships_deleted
    ON organization_team_memberships (deleted_at)
    WHERE status IN ('active', 'suspended');

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
    role varchar(30) NOT NULL DEFAULT 'member',
    status varchar(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'revoked', 'expired')),
    invited_by_mem_id uuid NOT NULL REFERENCES organization_memberships(id),
    token_hash varchar(255) UNIQUE NOT NULL,
    expires_at timestamptz NOT NULL,
    responded_at timestamptz,
    response varchar(30) CHECK (response IN ('accept', 'decline')),
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),

    -- Audit fields (who performed actions - stored in audit log, but useful here for quick access)
    responded_by_user_id uuid,  -- References external users table
    revoked_by_mem_id uuid REFERENCES organization_memberships(id)
);

COMMENT ON TABLE organization_invitations IS 'Tracks pending invitations to join an organization';
COMMENT ON COLUMN organization_invitations.token_hash IS 'SHA256 hash of the invitation token. Never store raw tokens.';
COMMENT ON COLUMN organization_invitations.expires_at IS 'Invitations are invalid after this timestamp (typically 7 days)';
COMMENT ON COLUMN organization_invitations.status IS 'pending=waiting, accepted=user joined, declined=user refused, revoked=admin cancelled, expired=auto after expires_at';

-- Use case: List invitations for an organization in member-management screens.
CREATE INDEX idx_invitations_organization_id ON organization_invitations (organization_id);
-- Use case: Find invitations sent to an email address during signup or invite acceptance.
CREATE INDEX idx_invitations_email ON organization_invitations (email);
-- Use case: Resolve an invitation by its token hash without scanning invitation records.
CREATE INDEX idx_invitations_token_hash ON organization_invitations (token_hash);
-- Use case: Find pending invitations due to expire or be marked expired.
CREATE INDEX idx_invitations_status_expires ON organization_invitations (status, expires_at)
    WHERE status = 'pending';

-- =====================================================
-- Organization Audit Logs (compliance & debugging)
-- =====================================================
CREATE TABLE organization_audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_mem_id uuid REFERENCES organization_memberships(id),  -- NULL for system actions
    action varchar(100) NOT NULL,
    target_type varchar(50),
    target_id uuid,
    old_value jsonb,
    new_value jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE organization_audit_logs IS 'Immutable audit trail for security, compliance, and debugging';
COMMENT ON COLUMN organization_audit_logs.actor_mem_id IS 'Membership ID of who performed action. NULL = system action (cleanup, migration)';
COMMENT ON COLUMN organization_audit_logs.action IS 'Verb like "member.added", "role.changed", "team.archived"';
COMMENT ON COLUMN organization_audit_logs.target_type IS 'Entity type: "organization", "membership", "team", "invitation"';
COMMENT ON COLUMN organization_audit_logs.old_value IS 'Previous state (for updates/changes)';
COMMENT ON COLUMN organization_audit_logs.new_value IS 'New state (for updates/changes)';
COMMENT ON COLUMN organization_audit_logs.metadata IS 'Additional context like {reason: "violation", source: "admin_api"}';

-- Use case: Read an organization's audit timeline in reverse chronological order.
CREATE INDEX idx_audit_org_time ON organization_audit_logs (organization_id, created_at DESC);
-- Use case: Investigate all audit events performed by a specific organization member.
CREATE INDEX idx_audit_actor ON organization_audit_logs (actor_mem_id, created_at DESC) WHERE actor_mem_id IS NOT NULL;
-- Use case: Read the audit history for a specific target entity.
CREATE INDEX idx_audit_target ON organization_audit_logs (target_type, target_id, created_at DESC) WHERE target_id IS NOT NULL;
-- Use case: Filter an organization's audit timeline by action type.
CREATE INDEX idx_audit_action ON organization_audit_logs (organization_id, action, created_at DESC);
-- Use case: Query recent audit events globally for monitoring and investigations.
CREATE INDEX idx_audit_created_at ON organization_audit_logs (created_at DESC);
