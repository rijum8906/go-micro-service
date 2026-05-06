package coreopenfga_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	taskpermissions "github.com/rijum8906/relay/packages/core/permissions/task"
)

func TestTaskModelProjectInheritanceIntegration(t *testing.T) {
	url := os.Getenv("OPENFGA_TEST_URL")
	if url == "" {
		t.Skip("OPENFGA_TEST_URL is not set")
	}

	ctx := context.Background()
	fgaClient, appErr := coreopenfga.NewClient(url)
	if appErr != nil {
		t.Fatal(appErr)
	}

	storeManager := coreopenfga.NewStoreManager(fgaClient.Client)
	store, appErr := storeManager.Create(ctx, fmt.Sprintf("task-model-%d", time.Now().UnixNano()))
	if appErr != nil {
		t.Fatal(appErr)
	}
	t.Cleanup(func() {
		if appErr := storeManager.Delete(context.Background()); appErr != nil {
			t.Logf("failed to delete OpenFGA test store %s: %v", store.Id, appErr)
		}
	})

	fgaClient.StoreID = storeManager.GetStoreID()
	modelManager := coreopenfga.NewModelManager(fgaClient.Client, storeManager)
	if appErr := modelManager.Write(ctx, "task"); appErr != nil {
		t.Fatal(appErr)
	}
	fgaClient.AuthorizationModelID = modelManager.GetAuthorizationModelID()

	tupleManager := coreopenfga.NewTupleManager(fgaClient)
	const user = "user:alice"
	const project = "project:project-1"
	const task = "task:task-1"
	if appErr := tupleManager.Write(ctx, []client.ClientTupleKey{
		{User: user, Relation: string(taskpermissions.RoleOwner), Object: project},
		{User: project, Relation: "parent_project", Object: task},
	}); appErr != nil {
		t.Fatal(appErr)
	}

	for _, relation := range []string{
		taskpermissions.PermissionCanView,
		taskpermissions.PermissionCanManage,
		taskpermissions.PermissionCanDelete,
	} {
		res, appErr := tupleManager.Check(ctx, client.ClientCheckRequest{
			User:     user,
			Relation: relation,
			Object:   task,
		})
		if appErr != nil {
			t.Fatal(appErr)
		}
		if res == nil || !res.GetAllowed() {
			t.Fatalf("expected %s to be allowed through project ownership", relation)
		}
	}
}
