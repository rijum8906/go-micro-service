package coreopenfga_test

import (
	"context"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
)

func mustCreateFGAClient() *client.OpenFgaClient {
	client, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: "http://localhost:9000",
	})
	if err != nil {
		panic(err)
	}
	return client
}

func Test_modelManager(t *testing.T) {
	fgaClient := mustCreateFGAClient()
	storeManager := coreopenfga.NewStoreManager(fgaClient)
	modelManager := coreopenfga.NewModelManager(fgaClient, storeManager)

	ctx := context.Background()

	_, err := storeManager.Create(ctx, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	// Write Auth Model
	err = modelManager.Write(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}

	// Get Auth Model
	_, err = modelManager.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		storeManager.Delete(context.Background())
	})
}
