package models_test

import (
	"context"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
)

// NOTE: Test_organizationModel_Permissions_ByRole tests all the permission that could be done directly
// - Organization[owner,admin,member] not any custom role
// - Team[leader,member] not any parent organiztion role
func Test_organizationModel_Permissions(t *testing.T) {
	fgaClient, err := coreopenfga.NewClient("http://localhost:9000")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	testStoreName := testutils.GenerateRandomString(10)

	storeManager := coreopenfga.NewStoreManager(fgaClient.Client)
	_, err = storeManager.Create(ctx, testStoreName)
	if err != nil {
		t.Fatal(err)
	}
	fgaClient.StoreID = storeManager.GetStoreID()
	modelManager := coreopenfga.NewModelManager(fgaClient.Client, storeManager)
	if err = modelManager.Write(ctx, "organization"); err != nil {
		t.Fatal(err)
	}
	fgaClient.AuthorizationModelID = modelManager.GetAuthorizationModelID()
	tupleManager := coreopenfga.NewTupleManager(fgaClient)

	testCases := []struct {
		name        string
		req         client.ClientTupleKey
		permissions []string
	}{
		{
			name: "Owner of organization should have all the permissions",
			req: client.ClientTupleKey{
				User:     "user:owner",
				Relation: permissions.RoleOwner,
				Object:   "organization:org-1",
			},
			permissions: []string{
				permissions.PermissionCanEdit,
				permissions.PermissionCanDelete,
				permissions.PermissionCanAddMember,
				permissions.PermissionCanRemoveMember,
				permissions.PermissionCanViewMember,
				permissions.PermissionCanCreateTeam,
			},
		},
		{
			name: "Admin of organization should have add, remove and view permissions",
			req: client.ClientTupleKey{
				User:     "user:admin",
				Relation: permissions.RoleAdmin,
				Object:   "organization:org-1",
			},
			permissions: []string{
				permissions.PermissionCanEdit,
				permissions.PermissionCanViewMember,
				permissions.PermissionCanAddMember,
				permissions.PermissionCanRemoveMember,
			},
		},
		{
			name: "Member should have view permissions and create team permission",
			req: client.ClientTupleKey{
				User:     "user:member",
				Relation: permissions.RoleMember,
				Object:   "organization:org-1",
			},
			permissions: []string{
				permissions.PermissionCanCreateTeam,
				permissions.PermissionCanViewMember,
			},
		},
		{
			name: "Leader should have all the permissions",
			req: client.ClientTupleKey{
				User:     "user:leader",
				Relation: permissions.RoleLeader,
				Object:   "team:backend",
			},
			permissions: []string{
				permissions.PermissionCanEdit,
				permissions.PermissionCanDelete,
				permissions.PermissionCanAddMember,
				permissions.PermissionCanRemoveMember,
				permissions.PermissionCanViewMember,
			},
		},
		{
			name: "Member should have all the permissions",
			req: client.ClientTupleKey{
				User:     "user:member",
				Relation: permissions.RoleMember,
				Object:   "team:backend",
			},
			permissions: []string{
				permissions.PermissionCanViewMember,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err = tupleManager.Write(ctx, []client.ClientTupleKey{testCase.req}); err != nil {
				t.Fatal(err)
			}

			for _, permission := range testCase.permissions {
				check, err := tupleManager.Check(ctx, client.ClientCheckRequest{
					User:     testCase.req.User,
					Relation: permission,
					Object:   testCase.req.Object,
				})
				if err != nil {
					t.Fatal(err)
				}
				if !*check.Allowed {
					t.Errorf("expected %s to be allowed", permission)
				}
			}

			t.Cleanup(func() {
				if err = tupleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
					{
						User:     testCase.req.User,
						Relation: testCase.req.Relation,
						Object:   testCase.req.Object,
					},
				}); err != nil {
					t.Fatal(err)
				}
			})
		})
	}

	t.Cleanup(func() {
		storeManager.Delete(context.Background())
	})
}
