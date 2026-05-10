// Package models_test provides comprehensive tests for the OpenFGA organization model.
// It covers system roles (owner, admin, member), custom roles with dynamic permissions,
// role inheritance, team membership, and negative/boundary cases.
package models_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Global test state — initialised once in TestMain
// ---------------------------------------------------------------------------

var (
	fgaClient    *coreopenfga.Client
	storeManager coreopenfga.StoreManager
)

// ---------------------------------------------------------------------------
// TestMain — boots a fresh OpenFGA store + model for the whole test binary
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	f, err := coreopenfga.NewClient("http://localhost:9000")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create OpenFGA client:", err)
		os.Exit(1)
	}
	fgaClient = f

	ctx := context.Background()
	sm := coreopenfga.NewStoreManager(f.Client)
	if _, err = sm.Create(ctx, testutils.GenerateRandomString(10)); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create store:", err)
		os.Exit(1)
	}
	storeManager = sm
	f.StoreID = sm.GetStoreID()

	mm := coreopenfga.NewModelManager(f.Client, sm)
	if err = mm.Write(ctx, "organization"); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write model:", err)
		os.Exit(1)
	}
	f.AuthorizationModelID = mm.GetAuthorizationModelID()

	code := m.Run()

	// Best-effort cleanup — ignore errors so a failed store delete
	// doesn't mask a real test failure.
	_ = storeManager.Delete(ctx)
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ids is a small convenience struct to hold random IDs for one test scenario.
type ids struct {
	userID string
	orgID  string
	teamID string
	roleID string
}

func newIDs() ids {
	return ids{
		userID: uuid.NewString(),
		orgID:  uuid.NewString(),
		teamID: uuid.NewString(),
		roleID: uuid.NewString(),
	}
}

// checkAll asserts that every (relation, object) pair in checks is allowed/denied
// for the given user.
type checkCase struct {
	relation string
	object   string
	want     bool   // true = allowed, false = denied
	label    string // human-readable description
}

func runChecks(
	t *testing.T,
	ctx context.Context,
	tm coreopenfga.TuppleManager,
	user string,
	cases []checkCase,
) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			result, appErr := tm.Check(ctx, client.ClientCheckRequest{
				User:     user,
				Relation: c.relation,
				Object:   c.object,
			})
			require.Nil(t, appErr, "Check() returned error for relation=%s", c.relation)
			require.NotNil(t, result)
			assert.Equal(t, c.want, *result.Allowed,
				"user=%s relation=%s object=%s", user, c.relation, c.object)
		})
	}
}

// writeAndRequire writes tuples and fails the test immediately on error.
func writeAndRequire(t *testing.T, ctx context.Context, tm coreopenfga.TuppleManager, tuples []client.ClientTupleKey) {
	t.Helper()
	require.Nil(t, tm.Write(ctx, tuples), "failed to write tuples")
}

// ---------------------------------------------------------------------------
// 1. Owner permissions
// ---------------------------------------------------------------------------

func Test_Owner_HasAllOrgPermissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: fmt.Sprintf("user:%s", i.userID), Relation: "owner", Object: fmt.Sprintf("organization:%s", i.orgID)},
	})

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)

	cases := []checkCase{
		{permissions.PermissionCanEditOrganization, org, true, "can_edit_organization"},
		{permissions.PermissionCanDeleteOrganization, org, true, "can_delete_organization"},
		{permissions.PermissionCanViewOrganization, org, true, "can_view_organization"},
		{permissions.PermissionCanTransferOwnership, org, true, "can_transfer_ownership"},
		{permissions.PermissionCanAddMember, org, true, "can_add_member"},
		{permissions.PermissionCanRemoveMember, org, true, "can_remove_member"},
		{permissions.PermissionCanViewMembers, org, true, "can_view_members"},
		{permissions.PermissionCanChangeMemberRole, org, true, "can_change_member_role"},
		{permissions.PermissionCanChangeMemberStatus, org, true, "can_change_member_status"},
		{permissions.PermissionCanInviteMember, org, true, "can_invite_member"},
		{permissions.PermissionCanApproveJoinRequest, org, true, "can_approve_join_request"},
		{permissions.PermissionCanCreateTeam, org, true, "can_create_team"},
		{permissions.PermissionCanDeleteTeam, org, true, "can_delete_team"},
		{permissions.PermissionCanViewTeams, org, true, "can_view_teams"},
		{permissions.PermissionCanCreateRole, org, true, "can_create_role"},
		{permissions.PermissionCanEditRole, org, true, "can_edit_role"},
		{permissions.PermissionCanDeleteRole, org, true, "can_delete_role"},
		{permissions.PermissionCanAssignRole, org, true, "can_assign_role"},
	}
	runChecks(t, ctx, tm, user, cases)
}

// ---------------------------------------------------------------------------
// 2. Admin permissions
// ---------------------------------------------------------------------------

func Test_Admin_HasAdminPermissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: fmt.Sprintf("user:%s", i.userID), Relation: "admin", Object: fmt.Sprintf("organization:%s", i.orgID)},
	})

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)

	cases := []checkCase{
		// Allowed
		{permissions.PermissionCanViewOrganization, org, true, "can_view_organization"},
		{permissions.PermissionCanAddMember, org, true, "can_add_member"},
		{permissions.PermissionCanRemoveMember, org, true, "can_remove_member"},
		{permissions.PermissionCanViewMembers, org, true, "can_view_members"},
		{permissions.PermissionCanChangeMemberRole, org, true, "can_change_member_role"},
		{permissions.PermissionCanChangeMemberStatus, org, true, "can_change_member_status"},
		{permissions.PermissionCanInviteMember, org, true, "can_invite_member"},
		{permissions.PermissionCanApproveJoinRequest, org, true, "can_approve_join_request"},
		{permissions.PermissionCanCreateTeam, org, true, "can_create_team"},
		{permissions.PermissionCanViewTeams, org, true, "can_view_teams"},
		{permissions.PermissionCanCreateRole, org, true, "can_create_role"},
		{permissions.PermissionCanAssignRole, org, true, "can_assign_role"},
		{permissions.PermissionCanDeleteTeam, org, true, "can_delete_team"},
		// Denied — owner-only
		{permissions.PermissionCanDeleteOrganization, org, false, "cannot_delete_organization"},
		{permissions.PermissionCanTransferOwnership, org, false, "cannot_transfer_ownership"},
	}
	runChecks(t, ctx, tm, user, cases)
}

// ---------------------------------------------------------------------------
// 3. Member permissions
// ---------------------------------------------------------------------------

func Test_Member_HasLimitedPermissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: fmt.Sprintf("user:%s", i.userID), Relation: "member", Object: fmt.Sprintf("organization:%s", i.orgID)},
	})

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)

	cases := []checkCase{
		// Allowed
		{permissions.PermissionCanViewMembers, org, true, "can_view_members"},
		{permissions.PermissionCanViewTeams, org, true, "can_view_teams"},
		// Denied
		{permissions.PermissionCanViewOrganization, org, true, "can_view_organization"},
		{permissions.PermissionCanEditOrganization, org, false, "cannot_edit_organization"},
		{permissions.PermissionCanDeleteOrganization, org, false, "cannot_delete_organization"},
		{permissions.PermissionCanTransferOwnership, org, false, "cannot_transfer_ownership"},
		{permissions.PermissionCanAddMember, org, false, "cannot_add_member"},
		{permissions.PermissionCanRemoveMember, org, false, "cannot_remove_member"},
		{permissions.PermissionCanChangeMemberRole, org, false, "cannot_change_member_role"},
		{permissions.PermissionCanChangeMemberStatus, org, false, "cannot_change_member_status"},
		{permissions.PermissionCanCreateTeam, org, false, "cannot_create_team"},
		{permissions.PermissionCanDeleteTeam, org, false, "cannot_delete_team"},
		{permissions.PermissionCanCreateRole, org, false, "cannot_create_role"},
		{permissions.PermissionCanAssignRole, org, false, "cannot_assign_role"},
	}
	runChecks(t, ctx, tm, user, cases)
}

// ---------------------------------------------------------------------------
// 4. Stranger — no relation at all
// ---------------------------------------------------------------------------

func Test_Stranger_HasNoPermissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	// Deliberately write NO tuples for this user.
	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)

	for _, perm := range permissions.OrgPermissions {
		t.Run("denied_"+perm, func(t *testing.T) {
			result, appErr := tm.Check(ctx, client.ClientCheckRequest{
				User:     user,
				Relation: perm,
				Object:   org,
			})
			require.Nil(t, appErr)
			require.NotNil(t, result)
			assert.False(t, *result.Allowed, "stranger should not have %s", perm)
		})
	}
}

// ---------------------------------------------------------------------------
// 5. Custom role — single permission
// ---------------------------------------------------------------------------

func Test_CustomRole_SinglePermission(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)
	role := fmt.Sprintf("role:%s", i.roleID)
	permObj := permissions.PermissionObject(i.orgID, permissions.PermissionCanViewMembers)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		// Role belongs to org
		{User: org, Relation: "organization", Object: role},
		// Grant permission to role
		{User: role, Relation: "granted_to", Object: permObj},
		// Assign user to role
		{User: user, Relation: "assignee", Object: role},
	})

	t.Run("granted_permission_allowed", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User:     user,
			Relation: "allowed",
			Object:   permObj,
		})
		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.True(t, *result.Allowed)
	})

	t.Run("ungranted_permission_denied", func(t *testing.T) {
		otherPerm := permissions.PermissionObject(i.orgID, permissions.PermissionCanDeleteOrganization)
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User:     user,
			Relation: "allowed",
			Object:   otherPerm,
		})
		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.False(t, *result.Allowed)
	})
}

// ---------------------------------------------------------------------------
// 6. Custom role — multiple permissions
// ---------------------------------------------------------------------------

func Test_CustomRole_MultiplePermissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)
	role := fmt.Sprintf("role:%s", i.roleID)

	granted := []string{
		permissions.PermissionCanViewMembers,
		permissions.PermissionCanInviteMember,
		permissions.PermissionCanViewTeams,
	}
	denied := []string{
		permissions.PermissionCanDeleteOrganization,
		permissions.PermissionCanTransferOwnership,
		permissions.PermissionCanCreateRole,
	}

	tuples := []client.ClientTupleKey{
		{User: org, Relation: "organization", Object: role},
		{User: user, Relation: "assignee", Object: role},
	}
	for _, p := range granted {
		tuples = append(tuples, client.ClientTupleKey{
			User:     role,
			Relation: "granted_to",
			Object:   permissions.PermissionObject(i.orgID, p),
		})
	}
	writeAndRequire(t, ctx, tm, tuples)

	for _, p := range granted {
		t.Run("granted_"+p, func(t *testing.T) {
			result, appErr := tm.Check(ctx, client.ClientCheckRequest{
				User:     user,
				Relation: "allowed",
				Object:   permissions.PermissionObject(i.orgID, p),
			})
			require.Nil(t, appErr)
			assert.True(t, *result.Allowed)
		})
	}

	for _, p := range denied {
		t.Run("denied_"+p, func(t *testing.T) {
			result, appErr := tm.Check(ctx, client.ClientCheckRequest{
				User:     user,
				Relation: "allowed",
				Object:   permissions.PermissionObject(i.orgID, p),
			})
			require.Nil(t, appErr)
			assert.False(t, *result.Allowed)
		})
	}
}

// ---------------------------------------------------------------------------
// 7. Role inheritance — child role inherits parent's permissions
// ---------------------------------------------------------------------------

func Test_CustomRole_Inheritance(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)

	parentRoleID := uuid.NewString()
	childRoleID := uuid.NewString()
	parentRole := fmt.Sprintf("role:%s", parentRoleID)
	childRole := fmt.Sprintf("role:%s", childRoleID)

	parentPermObj := permissions.PermissionObject(i.orgID, permissions.PermissionCanViewMembers)
	childPermObj := permissions.PermissionObject(i.orgID, permissions.PermissionCanInviteMember)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		// Roles belong to org
		{User: org, Relation: "organization", Object: parentRole},
		{User: org, Relation: "organization", Object: childRole},
		// Parent has one permission
		{User: parentRole, Relation: "granted_to", Object: parentPermObj},
		// Child has a different permission
		{User: childRole, Relation: "granted_to", Object: childPermObj},
		// Child inherits parent
		{User: childRole, Relation: "inherits", Object: parentRole},
		// User is assignee of child role only
		{User: user, Relation: "assignee", Object: childRole},
	})

	t.Run("child_own_permission_allowed", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "allowed", Object: childPermObj,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed)
	})

	t.Run("inherited_parent_permission_allowed", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "allowed", Object: parentPermObj,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed, "user should inherit parent role's permission")
	})
}

// ---------------------------------------------------------------------------
// 8. Team membership grants org membership
// ---------------------------------------------------------------------------

func Test_TeamMember_IsOrgMember(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	team := fmt.Sprintf("team:%s", i.teamID)
	user := fmt.Sprintf("user:%s", i.userID)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		// Team belongs to org
		{User: org, Relation: "organization", Object: team},
		// User is a team member
		{User: user, Relation: "member", Object: team},
		{
			User:     fmt.Sprintf("team:%s#member", i.teamID),
			Relation: "member",
			Object:   org,
		},
	})

	t.Run("team_member_can_view_organization", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: permissions.PermissionCanViewOrganization, Object: org,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed, "team member should be able to view org")
	})

	t.Run("team_member_cannot_edit_organization", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: permissions.PermissionCanEditOrganization, Object: org,
		})
		require.Nil(t, appErr)
		assert.False(t, *result.Allowed)
	})
}

// ---------------------------------------------------------------------------
// 9. Team permissions
// ---------------------------------------------------------------------------

func Test_Team_Permissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	ownerID := uuid.NewString()
	adminID := uuid.NewString()
	memberID := uuid.NewString()
	strangerID := uuid.NewString()

	org := fmt.Sprintf("organization:%s", i.orgID)
	team := fmt.Sprintf("team:%s", i.teamID)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: org, Relation: "organization", Object: team},
		{User: fmt.Sprintf("user:%s", ownerID), Relation: "owner", Object: team},
		{User: fmt.Sprintf("user:%s", adminID), Relation: "admin", Object: team},
		{User: fmt.Sprintf("user:%s", memberID), Relation: "member", Object: team},
	})

	tests := []struct {
		name     string
		userID   string
		relation string
		want     bool
	}{
		// Owner
		{"owner_can_edit_team", ownerID, "can_edit_team", true},
		{"owner_can_delete_team", ownerID, "can_delete_team", true},
		{"owner_can_add_member", ownerID, "can_add_member", true},
		{"owner_can_view_team", ownerID, "can_view_team", true},
		// Admin
		{"admin_can_edit_team", adminID, "can_edit_team", true},
		{"admin_can_add_member", adminID, "can_add_member", true},
		{"admin_can_view_team", adminID, "can_view_team", true},
		{"admin_cannot_delete_team", adminID, "can_delete_team", false},
		// Member
		{"member_can_view_team", memberID, "can_view_team", true},
		{"member_cannot_edit_team", memberID, "can_edit_team", false},
		{"member_cannot_add_member", memberID, "can_add_member", false},
		// Stranger
		{"stranger_cannot_view_team", strangerID, "can_view_team", false},
		{"stranger_cannot_edit_team", strangerID, "can_edit_team", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, appErr := tm.Check(ctx, client.ClientCheckRequest{
				User:     fmt.Sprintf("user:%s", tt.userID),
				Relation: tt.relation,
				Object:   team,
			})
			require.Nil(t, appErr)
			require.NotNil(t, result)
			assert.Equal(t, tt.want, *result.Allowed)
		})
	}
}

// ---------------------------------------------------------------------------
// 10. Org admin is automatically a team admin
// ---------------------------------------------------------------------------

func Test_OrgAdmin_IsTeamAdmin(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	team := fmt.Sprintf("team:%s", i.teamID)
	user := fmt.Sprintf("user:%s", i.userID)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		// User is org admin — NOT explicitly a team admin
		{User: user, Relation: "admin", Object: org},
		{User: org, Relation: "organization", Object: team},
	})

	t.Run("org_admin_can_edit_team", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "can_edit_team", Object: team,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed, "org admin should implicitly be team admin")
	})

	t.Run("org_admin_can_add_team_member", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "can_add_member", Object: team,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed)
	})
}

// ---------------------------------------------------------------------------
// 11. Permission isolation across organizations
// ---------------------------------------------------------------------------

func Test_Permission_IsolatedAcrossOrgs(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()

	userID := uuid.NewString()
	org1ID := uuid.NewString()
	org2ID := uuid.NewString()
	roleID := uuid.NewString()

	user := fmt.Sprintf("user:%s", userID)
	org1 := fmt.Sprintf("organization:%s", org1ID)
	role := fmt.Sprintf("role:%s", roleID)

	// Grant permission only in org1
	permObj1 := permissions.PermissionObject(org1ID, permissions.PermissionCanViewMembers)
	permObj2 := permissions.PermissionObject(org2ID, permissions.PermissionCanViewMembers)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: org1, Relation: "organization", Object: role},
		{User: role, Relation: "granted_to", Object: permObj1},
		{User: user, Relation: "assignee", Object: role},
	})

	t.Run("allowed_in_org1", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "allowed", Object: permObj1,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed)
	})

	t.Run("denied_in_org2", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "allowed", Object: permObj2,
		})
		require.Nil(t, appErr)
		assert.False(t, *result.Allowed, "permission should not bleed into org2")
	})
}

// ---------------------------------------------------------------------------
// 12. Role assigned to multiple users
// ---------------------------------------------------------------------------

func Test_CustomRole_MultipleAssignees(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	user1ID := uuid.NewString()
	user2ID := uuid.NewString()
	user3ID := uuid.NewString() // not assigned

	org := fmt.Sprintf("organization:%s", i.orgID)
	role := fmt.Sprintf("role:%s", i.roleID)
	permObj := permissions.PermissionObject(i.orgID, permissions.PermissionCanInviteMember)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: org, Relation: "organization", Object: role},
		{User: role, Relation: "granted_to", Object: permObj},
		{User: fmt.Sprintf("user:%s", user1ID), Relation: "assignee", Object: role},
		{User: fmt.Sprintf("user:%s", user2ID), Relation: "assignee", Object: role},
	})

	for _, uid := range []string{user1ID, user2ID} {
		t.Run("assigned_user_"+uid[:8]+"_allowed", func(t *testing.T) {
			result, appErr := tm.Check(ctx, client.ClientCheckRequest{
				User: fmt.Sprintf("user:%s", uid), Relation: "allowed", Object: permObj,
			})
			require.Nil(t, appErr)
			assert.True(t, *result.Allowed)
		})
	}

	t.Run("unassigned_user_denied", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: fmt.Sprintf("user:%s", user3ID), Relation: "allowed", Object: permObj,
		})
		require.Nil(t, appErr)
		assert.False(t, *result.Allowed)
	})
}

// ---------------------------------------------------------------------------
// 13. User assigned to multiple roles — union of permissions
// ---------------------------------------------------------------------------

func Test_User_MultipleRoles_UnionOfPermissions(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	roleAID := uuid.NewString()
	roleBID := uuid.NewString()

	org := fmt.Sprintf("organization:%s", i.orgID)
	user := fmt.Sprintf("user:%s", i.userID)
	roleA := fmt.Sprintf("role:%s", roleAID)
	roleB := fmt.Sprintf("role:%s", roleBID)

	permA := permissions.PermissionObject(i.orgID, permissions.PermissionCanViewMembers)
	permB := permissions.PermissionObject(i.orgID, permissions.PermissionCanCreateTeam)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: org, Relation: "organization", Object: roleA},
		{User: org, Relation: "organization", Object: roleB},
		{User: roleA, Relation: "granted_to", Object: permA},
		{User: roleB, Relation: "granted_to", Object: permB},
		{User: user, Relation: "assignee", Object: roleA},
		{User: user, Relation: "assignee", Object: roleB},
	})

	t.Run("permission_from_roleA", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "allowed", Object: permA,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed)
	})

	t.Run("permission_from_roleB", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: "allowed", Object: permB,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed)
	})
}

// ---------------------------------------------------------------------------
// 14. Org owner is automatically team owner / can delete team
// ---------------------------------------------------------------------------

func Test_OrgOwner_CanDeleteTeam(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	team := fmt.Sprintf("team:%s", i.teamID)
	user := fmt.Sprintf("user:%s", i.userID)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: user, Relation: "owner", Object: org},
		{User: org, Relation: "organization", Object: team},
	})

	result, appErr := tm.Check(ctx, client.ClientCheckRequest{
		User: user, Relation: "can_delete_team", Object: team,
	})
	require.Nil(t, appErr)
	assert.True(t, *result.Allowed, "org owner should be able to delete any team")
}

// ---------------------------------------------------------------------------
// 15. Role-based membership grants org member permissions
// ---------------------------------------------------------------------------

func Test_RoleAssignee_IsOrgMember(t *testing.T) {
	tm := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	i := newIDs()

	org := fmt.Sprintf("organization:%s", i.orgID)
	role := fmt.Sprintf("role:%s", i.roleID)
	user := fmt.Sprintf("user:%s", i.userID)

	writeAndRequire(t, ctx, tm, []client.ClientTupleKey{
		{User: org, Relation: "organization", Object: role},
		// role#assignee is a member of the org
		{User: fmt.Sprintf("role:%s#assignee", i.roleID), Relation: "member", Object: org},
		{User: user, Relation: "assignee", Object: role},
	})

	t.Run("role_assignee_can_view_organization", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: permissions.PermissionCanViewOrganization, Object: org,
		})
		require.Nil(t, appErr)
		assert.True(t, *result.Allowed)
	})

	t.Run("role_assignee_cannot_edit_organization", func(t *testing.T) {
		result, appErr := tm.Check(ctx, client.ClientCheckRequest{
			User: user, Relation: permissions.PermissionCanEditOrganization, Object: org,
		})
		require.Nil(t, appErr)
		assert.False(t, *result.Allowed)
	})
}
