package coreopenfga_test

import (
	"context"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/testutils"
)

func Test_tupleManager(t *testing.T) {
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
	if err = modelManager.Write(ctx, "test"); err != nil {
		t.Fatal(err)
	}
	fgaClient.AuthorizationModelID = modelManager.GetAuthorizationModelID()
	tupleManager := coreopenfga.NewTupleManager(fgaClient)

	// create tuple
	if appErr := tupleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:john",
			Relation: "author",
			Object:   "book:go.pdf",
		},
	}); appErr != nil {
		t.Fatal(err)
	}

	// read tuple
	res, err := tupleManager.Read(ctx, client.ClientReadRequest{
		User:     openfga.PtrString("user:john"),
		Relation: openfga.PtrString("author"),
		Object:   openfga.PtrString("book:go.pdf"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Tuples[0].Key.User != "user:john" {
		t.Errorf("expected user to be user:john but got %s", res.Tuples[0].Key.User)
	}
	if res.Tuples[0].Key.Relation != "author" {
		t.Errorf("expected relation to be admin but got %s", res.Tuples[0].Key.Relation)
	}
	if res.Tuples[0].Key.Object != "book:go.pdf" {
		t.Errorf("expected object to be organization:org-1 but got %s", res.Tuples[0].Key.Object)
	}

	// check tuple
	check, err := tupleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:john",
		Relation: "author",
		Object:   "book:go.pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !*check.Allowed {
		t.Errorf("expected allowed to be true but got %v", *check.Allowed)
	}

	// delete tuple
	if err = tupleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "user:john",
			Relation: "author",
			Object:   "book:go.pdf",
		},
	}); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		storeManager.Delete(context.Background())
	})
}
