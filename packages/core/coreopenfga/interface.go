package coreopenfga

import (
	"context"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Client struct {
	StoreID              string
	AuthorizationModelID string
	Client               *client.OpenFgaClient
}

type StoreManager interface {
	Create(ctx context.Context, name string) (*client.ClientCreateStoreResponse, error)
	Get(ctx context.Context) (*client.ClientGetStoreResponse, error)
	List(ctx context.Context) (*client.ClientListStoresResponse, error)
	Delete(ctx context.Context) error
	GetStoreID() string
	SetStoreID(storeID string)
}

type ModelManager interface {
	Write(ctx context.Context, modelName string) error
	Read(ctx context.Context) (*client.ClientReadAuthorizationModelResponse, error)
	GetAuthorizationModelID() string
	SetAuthorizationModelID(id string)
}

type TuppleManager interface {
	Write(ctx context.Context, writes []client.ClientTupleKey) error
	Read(ctx context.Context, req client.ClientReadRequest) (*client.ClientReadResponse, error)
	Delete(ctx context.Context, deletes []client.ClientTupleKeyWithoutCondition) error
}

func NewClient(url string) (*Client, error) {
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: url,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create FGA client").WithDetail("error", err.Error())
	}

	return &Client{
		Client: fgaClient,
	}, nil
}
