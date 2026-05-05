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
	Create(ctx context.Context, name string) (*client.ClientCreateStoreResponse, *apperror.AppError)
	Get(ctx context.Context) (*client.ClientGetStoreResponse, *apperror.AppError)
	List(ctx context.Context) (*client.ClientListStoresResponse, *apperror.AppError)
	Delete(ctx context.Context) *apperror.AppError
	GetStoreID() string
	SetStoreID(storeID string)
}

type ModelManager interface {
	Write(ctx context.Context, modelName string) *apperror.AppError
	Read(ctx context.Context) (*client.ClientReadAuthorizationModelResponse, *apperror.AppError)
	GetAuthorizationModelID() string
	SetAuthorizationModelID(id string)
}

type TuppleManager interface {
	Write(ctx context.Context, writes []client.ClientTupleKey) *apperror.AppError
	Read(ctx context.Context, req client.ClientReadRequest) (*client.ClientReadResponse, *apperror.AppError)
	Check(ctx context.Context, req client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError)
	Delete(ctx context.Context, deletes []client.ClientTupleKeyWithoutCondition) *apperror.AppError
}

func NewClient(url string) (*Client, *apperror.AppError) {
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
