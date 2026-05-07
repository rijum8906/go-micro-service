package testutils

import (
	"os"

	"github.com/rijum8906/relay/packages/core/coreopenfga"
)

func MustConnectFGAClient() *coreopenfga.Client {
	apiURL := os.Getenv("OPENFGA_TEST_API_URL")
	storeID := os.Getenv("OPENFGA_TEST_STORE_ID")
	modelID := os.Getenv("OPENFGA_TEST_AUTH_MODEL_ID")

	fgaClient, err := coreopenfga.NewClient(apiURL)
	if err != nil {
		panic(err)
	}

	fgaClient.StoreID = storeID
	fgaClient.AuthorizationModelID = modelID
	return fgaClient
}
