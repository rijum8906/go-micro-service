// Package coreopenfga
package coreopenfga

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
	"github.com/openfga/language/pkg/go/transformer"
)

type Client struct {
	FGAClient            *client.OpenFgaClient
	StoreID              string
	AuthorizationModelID string
}

type Config struct {
	APIURL    string // http://localhost:8080
	StoreName string // "relay_store"
	Token     string // Optional: for authentication
}

func NewClient(cfg Config) (*Client, error) {
	apiClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiScheme: "http",
		ApiHost:   cfg.APIURL,
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodNone,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return &Client{
		FGAClient: apiClient,
	}, nil
}

func (c *Client) CreateStore(ctx context.Context, name string) (string, error) {
	body := openfga.CreateStoreRequest{Name: name}
	createStoreRes, _, err := c.FGAClient.OpenFgaApi.CreateStore(ctx).Body(body).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create store: %w", err)
	}

	c.StoreID = createStoreRes.GetId()
	c.FGAClient.SetStoreId(c.StoreID)
	return c.StoreID, nil
}

func (c *Client) GetStore(ctx context.Context, name string) (string, error) {
	stores, err := c.FGAClient.ListStores(ctx).Execute()
	if err != nil {
		return "", err
	}

	for _, store := range stores.GetStores() {
		if store.GetName() == name {
			c.StoreID = store.GetId()
			return c.StoreID, nil
		}
	}

	return "", fmt.Errorf("store not found: %s", name)
}

func (c *Client) DeleteStore(ctx context.Context) error {
	_, err := c.FGAClient.OpenFgaApi.DeleteStore(ctx, c.StoreID).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete store: %w", err)
	}
	return nil
}

func (c *Client) WriteAuthModel(ctx context.Context) (*client.ClientWriteAuthorizationModelResponse, error) {
	if c.StoreID == "" {
		return nil, fmt.Errorf("store id is empty")
	}

	dslContent, _ := os.ReadFile("model.fga")
	jsonModel, _ := transformer.TransformDSLToJSON(string(dslContent))

	var body openfga.WriteAuthorizationModelRequest
	json.Unmarshal([]byte(jsonModel), &body)

	options := client.ClientWriteAuthorizationModelOptions{
		StoreId: openfga.PtrString(c.StoreID),
	}

	data, err := c.FGAClient.WriteAuthorizationModel(context.Background()).Options(options).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to write auth model: %w", err)
	}

	c.AuthorizationModelID = data.AuthorizationModelId

	return data, nil
}
