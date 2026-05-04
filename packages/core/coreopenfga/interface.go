package coreopenfga

import (
	"context"

	"github.com/openfga/go-sdk/client"
)

type Client interface {
	Connect()
	CreateStore()
	GetStore()
	ListStores()
	DeleteStore()
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
