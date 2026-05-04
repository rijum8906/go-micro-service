package coreopenfga_test

import (
	"context"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
)

func Test_storeManager(t *testing.T) {
	fgaClient := mustCreateFGAClient()
	storeManager := coreopenfga.NewStoreManager(fgaClient)

	ctx := context.Background()

	// Create a store and validatea it
	storeRes, err := storeManager.Create(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	if storeRes.Name != "test-store" {
		t.Errorf("expected store name to be test-store, got %s", storeRes.Name)
	}

	// Get the same store and validate
	storeRes2, err := storeManager.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if storeRes.Id != storeRes2.Id {
		t.Errorf("expected store id to be %s, got %s", storeRes.Id, storeRes2.Id)
	}

	// Delete Store
	err = storeManager.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Again try to get the store
	_, err = storeManager.Get(ctx)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	t.Cleanup(func() {
		if err == nil {
			fgaClient.DeleteStore(context.Background()).Options(client.ClientDeleteStoreOptions{
				StoreId: &storeRes.Id,
			}).Execute()
		}
	})
}
