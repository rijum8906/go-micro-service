package coreopenfga_test

import (
	"context"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
)

func TestNewClient(t *testing.T) {
	_, err := coreopenfga.NewClient(coreopenfga.Config{
		APIURL: "localhost:9000",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Client_CreateStore(t *testing.T) {
	fgaClient, err := coreopenfga.NewClient(coreopenfga.Config{
		APIURL: "localhost:9000",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	storeID, err := fgaClient.CreateStore(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	res, err := fgaClient.FGAClient.GetStore(ctx).Options(client.ClientGetStoreOptions{
		StoreId: &storeID,
	}).Execute()
	if err != nil {
		t.Fatal(err)
	}

	if res.Id != storeID {
		t.Fatalf("expected store id %s, got %s", storeID, res.Id)
	}

	t.Cleanup(func() {
		fgaClient.FGAClient.DeleteStore(ctx).Options(client.ClientDeleteStoreOptions{
			StoreId: &storeID,
		}).Execute()
	})
}

func Test_Client_GetStore(t *testing.T) {
	fgaClient, err := coreopenfga.NewClient(coreopenfga.Config{
		APIURL: "localhost:9000",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	storeID, err := fgaClient.CreateStore(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	id, err := fgaClient.GetStore(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	if id != storeID {
		t.Fatalf("expected store id %s, got %s", storeID, id)
	}

	t.Cleanup(func() {
		fgaClient.FGAClient.DeleteStore(ctx).Options(client.ClientDeleteStoreOptions{
			StoreId: &storeID,
		}).Execute()
	})
}

func Test_Client_DeleteStore(t *testing.T) {
	fgaClient, err := coreopenfga.NewClient(coreopenfga.Config{
		APIURL: "localhost:9000",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	storeID, err := fgaClient.CreateStore(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	err = fgaClient.DeleteStore(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fgaClient.FGAClient.GetStore(ctx).Options(client.ClientGetStoreOptions{
		StoreId: &storeID,
	}).Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Client_WriteAuthModel(t *testing.T) {
	fgaClient, err := coreopenfga.NewClient(coreopenfga.Config{
		APIURL: "localhost:9000",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	storeID, err := fgaClient.CreateStore(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fgaClient.WriteAuthModel(ctx)
	if err != nil {
		t.Fatal(err)
	}

	fgaClient.FGAClient.GetAuthorizationModelId()

	t.Cleanup(func() {
		fgaClient.FGAClient.DeleteStore(ctx).Options(client.ClientDeleteStoreOptions{
			StoreId: &storeID,
		}).Execute()
	})
}
