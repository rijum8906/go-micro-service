package coreopenfga

import (
	"context"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type storeManager struct {
	storeID   string
	fgaClient *client.OpenFgaClient
}

// Constructor

func NewStoreManager(fgaClient *client.OpenFgaClient) StoreManager {
	return &storeManager{
		fgaClient: fgaClient,
	}
}

// Methods

func (s *storeManager) GetStoreID() string {
	return s.storeID
}

func (s *storeManager) SetStoreID(storeID string) {
	s.storeID = storeID
}

func (s *storeManager) Create(ctx context.Context, name string) (*client.ClientCreateStoreResponse, error) {
	res, err := s.fgaClient.CreateStore(ctx).Body(client.ClientCreateStoreRequest{
		Name: name,
	}).Execute()
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to create store").WithDetail("error", err.Error())
	}
	s.SetStoreID(res.Id)

	return res, nil
}

func (s *storeManager) Get(ctx context.Context) (*client.ClientGetStoreResponse, error) {
	res, err := s.fgaClient.GetStore(ctx).Options(client.ClientGetStoreOptions{
		StoreId: &s.storeID,
	}).Execute()
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to get store").WithDetail("error", err.Error())
	}

	return res, nil
}

func (s *storeManager) List(ctx context.Context) (*client.ClientListStoresResponse, error) {
	res, err := s.fgaClient.ListStores(ctx).Execute()
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to list stores").WithDetail("error", err.Error())
	}

	return res, nil
}

func (s *storeManager) Delete(ctx context.Context) error {
	_, err := s.fgaClient.DeleteStore(ctx).Options(client.ClientDeleteStoreOptions{
		StoreId: &s.storeID,
	}).Execute()
	if err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to delete store").WithDetail("error", err.Error())
	}

	return nil
}
