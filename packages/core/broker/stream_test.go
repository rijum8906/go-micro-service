package broker_test

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/testutils"
)

func Test_StreamManager_Create(t *testing.T) {
	cfg := createStreamConfig()
	client := mustCreateClient()
	stream := broker.NewStreamManager(client.GetClient())

	streamInfo, appErr := stream.Create(cfg)
	if appErr != nil {
		t.Errorf("Create() failed with error: %v and details: %v", appErr, appErr.Details)
	}

	defer client.GetClient().JS.DeleteStream(streamInfo.Config.Name)
}

func Test_StreamManager_Delete(t *testing.T) {
	cfg := createStreamConfig()
	client := mustCreateClient()
	stream := broker.NewStreamManager(client.GetClient())

	streamInfo, appErr := stream.Create(cfg)
	if appErr != nil {
		t.Errorf("Create() failed with error: %v and details: %v", appErr, appErr.Details)
	}

	if appErr = stream.Delete(streamInfo.Config.Name); appErr != nil {
		t.Errorf("Delete() doesn't want error but got %v", appErr)
	}

	if stream.Exists(streamInfo.Config.Name) {
		t.Errorf("stream didn't delete successfully")
	}
}

func createStreamConfig() *broker.StreamConfig {
	return broker.NewStreamConfig(testutils.GenerateRandomString(5)).
		AddStorageType(nats.FileStorage)
}

func mustCreateClient() broker.Client {
	client := broker.NewClient()
	if appErr := client.Connect("http://localhost:4223"); appErr != nil {
		panic(appErr.Details)
	}

	return client
}
