package broker_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/testutils"
)

func Test_Consumer_Create(t *testing.T) {
	client := broker.NewClient()
	// Use a context with timeout for tests so they don't hang
	appErr := client.Connect("nats://localhost:4223")
	if appErr != nil {
		t.Fatalf("can't connect to client: %v", appErr)
	}

	streamManager := broker.NewStreamManager(client.GetClient())
	streamName := testutils.GenerateRandomString(5)
	subject := testutils.GenerateRandomString(5)
	streamCfg := broker.NewStreamConfig(streamName)
	streamCfg.AddSubjects(subject)

	// Create Stream
	_, appErr = streamManager.Create(streamCfg)
	if appErr != nil {
		t.Fatalf("can't create stream error: %v", appErr)
	}

	defer streamManager.Delete(streamName)

	consumerManager := broker.NewConsumerManager(client.GetClient())
	consumerName := testutils.GenerateRandomString(5)
	consumerCfg := broker.NewConsumerConfig(consumerName)

	_, appErr = consumerManager.Create(streamName, consumerCfg)
	if appErr != nil {
		t.Errorf("can't create consumer: %v", appErr)
	}

	defer consumerManager.Delete(streamName, consumerName)
}

func Test_Consumer_Delete(t *testing.T) {
	client := broker.NewClient()
	// Use a context with timeout for tests so they don't hang
	appErr := client.Connect("nats://localhost:4223")
	if appErr != nil {
		t.Fatalf("can't connect to client: %v", appErr)
	}

	streamManager := broker.NewStreamManager(client.GetClient())
	streamName := testutils.GenerateRandomString(5)
	subject := testutils.GenerateRandomString(5)
	streamCfg := broker.NewStreamConfig(streamName)
	streamCfg.AddSubjects(subject)

	// Create Stream
	_, appErr = streamManager.Create(streamCfg)
	if appErr != nil {
		t.Fatalf("can't create stream error: %v", appErr)
	}

	defer streamManager.Delete(streamName)

	consumerManager := broker.NewConsumerManager(client.GetClient())
	consumerName := testutils.GenerateRandomString(5)
	consumerCfg := broker.NewConsumerConfig(consumerName)

	_, appErr = consumerManager.Create(streamName, consumerCfg)
	if appErr != nil {
		t.Errorf("can't create consumer: %v", appErr)
	}

	if appErr = consumerManager.Delete(streamName, consumerName); appErr != nil {
		t.Errorf("can't delete consumer: %v", appErr)
	}

	_, appErr = consumerManager.Get(streamName, consumerName)
	if appErr == nil {
		t.Errorf("consumer still exists after delete")
	}
}
