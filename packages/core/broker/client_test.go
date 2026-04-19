package broker_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
)

func Test_Client_Connect_Success(t *testing.T) {
	client := broker.NewClient()
	appErr := client.Connect("http://localhost:4223")
	if appErr != nil {
		t.Errorf("Error connecting NATS. Details:%v", appErr.Details)
	}

	isConnected := client.IsConnected()
	if !isConnected {
		t.Errorf("IsConnected() showing %v wants %v", isConnected, true)
	}
}

func Test_Client_Connect_Failure(t *testing.T) {
	client := broker.NewClient()
	appErr := client.Connect("http://localhost:4222")
	if appErr == nil {
		t.Errorf("Connect() wants error with wrong address but got no error")
		return
	}

	if appErr.Code != apperror.CodeThirdParty {
		t.Errorf("Connect() wants the error code to be %v but got %v", apperror.CodeThirdParty, appErr.Code)
	}
}

func Test_Client_Drain(t *testing.T) {
	client := broker.NewClient()
	appErr := client.Connect("http://localhost:4223")
	if appErr != nil {
		t.Errorf("Error connecting NATS. Details:%v", appErr.Details)
	}

	if appErr = client.Drain(); appErr != nil {
		t.Errorf("Drain() is returning error %v, Details: %v", appErr, appErr.Details)
	}
}

func Test_Client_Close(t *testing.T) {
	client := broker.NewClient()
	appErr := client.Connect("http://localhost:4223")
	if appErr != nil {
		t.Errorf("Error connecting NATS. Details:%v", appErr.Details)
	}

	if appErr = client.Close(); appErr != nil {
		t.Errorf("Close() is returning error %v, Details: %v", appErr, appErr.Details)
	}
}
