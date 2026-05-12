package permissions_test

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
	"github.com/stretchr/testify/require"
)

var (
	fgaClient    *coreopenfga.Client
	storeManager coreopenfga.StoreManager
)

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

func Test_CreateCustomRole(t *testing.T) {
	permissionManager := permissions.NewPermissionManager(fgaClient)
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)
	ctx := context.Background()
	userID := uuid.NewString()
	orgID := uuid.NewString()
	role := "custom_role"

	appErr := permissionManager.CreateCustomRole(ctx, userID, orgID, role, permissions.PermissionCanChangeMemberStatus)
	require.Nil(t, appErr)

	res, appErr := tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userID,
		Relation: "allowed",
		Object:   permissions.GeneratePermissionObject(orgID, permissions.PermissionCanChangeMemberStatus),
	})
	require.Nil(t, appErr)
	require.True(t, *res.Allowed)
}
