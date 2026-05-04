package coreopenfga

import (
	"context"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type tupleManager struct {
	client *Client
}

// Constructor

func NewTupleManager(fgaClient *Client) *tupleManager {
	return &tupleManager{
		client: fgaClient,
	}
}

// Methods

func (m *tupleManager) Write(ctx context.Context, creates []client.ClientTupleKey) error {
	options := client.ClientWriteOptions{
		AuthorizationModelId: &m.client.AuthorizationModelID,
		StoreId:              &m.client.StoreID,
	}

	_, err := m.client.Client.Write(ctx).Body(client.ClientWriteRequest{
		Writes: creates,
	}).Options(options).Execute()
	if err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to write tuple").WithDetail("error", err.Error())
	}

	return nil
}

func (m *tupleManager) Read(ctx context.Context, body client.ClientReadRequest) (*client.ClientReadResponse, error) {
	options := client.ClientReadOptions{
		StoreId:  &m.client.StoreID,
		PageSize: openfga.PtrInt32(10),
	}

	res, err := m.client.Client.Read(ctx).Body(body).Options(options).Execute()
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to read tuple").WithDetail("error", err.Error())
	}

	return res, nil
}

func (m *tupleManager) Check(ctx context.Context, body client.ClientCheckRequest) (*client.ClientCheckResponse, error) {
	options := client.ClientCheckOptions{
		StoreId:              &m.client.StoreID,
		AuthorizationModelId: &m.client.AuthorizationModelID,
	}

	res, err := m.client.Client.Check(ctx).Body(body).Options(options).Execute()
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to check tuple").WithDetail("error", err.Error())
	}

	return res, nil
}

func (m *tupleManager) Delete(ctx context.Context, deletes []client.ClientTupleKeyWithoutCondition) error {
	options := client.ClientWriteOptions{
		AuthorizationModelId: &m.client.AuthorizationModelID,
		StoreId:              &m.client.StoreID,
	}

	_, err := m.client.Client.Write(ctx).Body(client.ClientWriteRequest{
		Deletes: deletes,
	}).Options(options).Execute()
	if err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to write tuple").WithDetail("error", err.Error())
	}

	return nil
}
