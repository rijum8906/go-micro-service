// Package coreopenfga
package coreopenfga

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/language/pkg/go/transformer"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type modelManager struct {
	fgaClient            *client.OpenFgaClient
	storeManager         StoreManager
	authorizationModelID string
}

// Constructior

func NewModelManager(fgaClient *client.OpenFgaClient, storeManager StoreManager) ModelManager {
	return &modelManager{
		fgaClient:    fgaClient,
		storeManager: storeManager,
	}
}

// Methods

func (m *modelManager) GetAuthorizationModelID() string {
	return m.authorizationModelID
}

func (m *modelManager) SetAuthorizationModelID(id string) {
	m.authorizationModelID = id
}

func (m *modelManager) Write(ctx context.Context, name string) *apperror.AppError {
	dslContent, err := os.ReadFile(fmt.Sprintf("./models/%s_model.fga", name))
	if err != nil {
		return apperror.New(apperror.CodeNotFound, "model not found").WithDetail("error", err.Error())
	}
	jsonModel, err := transformer.TransformDSLToJSON(string(dslContent))
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to transform dsl to json").WithDetail("error", err.Error())
	}

	var body openfga.WriteAuthorizationModelRequest
	if err := json.Unmarshal([]byte(jsonModel), &body); err != nil {
		return apperror.New(apperror.CodeInternal, "failed to parse model json").WithDetail("error", err.Error())
	}

	options := client.ClientWriteAuthorizationModelOptions{
		StoreId: openfga.PtrString(m.storeManager.GetStoreID()),
	}

	res, err := m.fgaClient.WriteAuthorizationModel(ctx).Options(options).Body(body).Execute()
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to write auth model").WithDetail("error", err.Error())
	}
	m.authorizationModelID = res.AuthorizationModelId

	return nil
}

func (m *modelManager) Read(ctx context.Context) (*client.ClientReadAuthorizationModelResponse, *apperror.AppError) {
	options := client.ClientReadAuthorizationModelOptions{
		AuthorizationModelId: openfga.PtrString(m.authorizationModelID),
		StoreId:              openfga.PtrString(m.storeManager.GetStoreID()),
	}

	res, err := m.fgaClient.ReadAuthorizationModel(ctx).Options(options).Body(client.ClientReadAuthorizationModelRequest{}).Execute()
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to get auth model").WithDetail("error", err.Error())
	}

	return res, nil
}
