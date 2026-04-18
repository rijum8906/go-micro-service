package broker_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
)

func TestClient_CreateStream(t *testing.T) {
	client := mustConnectClient()
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg     broker.StreamConfig
		wantErr bool
		errCode apperror.ErrorCode
	}{
		{
			name:    "invalid stream config",
			cfg:     broker.StreamConfig{},
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name: "missing name",
			cfg: broker.StreamConfig{
				Subjects:   []string{"test-subject"},
				MaxBytes:   10000000,
				MaxAge:     time.Hour,
				MaxMsgs:    1000,
				MaxMsgSize: 1024 * 1024,
				Storage:    nats.FileStorage,
				Replicas:   1,
				Duplicates: time.Minute,
				Retention:  nats.WorkQueuePolicy,
				Discard:    nats.DiscardNew,
			},
			wantErr: true,
			errCode: apperror.CodeValidation,
		},

		{
			name:    "valid stream config",
			cfg:     crateStreamConfig("test-stream", "test-subject"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := client.CreateStream(tt.cfg)
			if tt.wantErr {
				if appErr == nil {
					t.Fatalf("CreateStream() = nil, want error")
				}

				if appErr.Code != tt.errCode {
					t.Fatalf("CreateStream() = %v, want %v", appErr.Code, tt.errCode)
				}
			} else {
				if appErr != nil {
					t.Fatalf("CreateStream() = %v, want nil\nDetails:%v", appErr, appErr.Details)
				}
			}
		})
	}
}

func crateStreamConfig(name string, subject string) broker.StreamConfig {
	return broker.StreamConfig{
		Name:       name,
		Subjects:   []string{subject},
		MaxBytes:   10000000,
		MaxAge:     time.Hour,
		MaxMsgs:    1000,
		MaxMsgSize: 1024 * 1024,
		Storage:    nats.FileStorage,
		Replicas:   1,
		Duplicates: time.Minute,
		Retention:  nats.WorkQueuePolicy,
		Discard:    nats.DiscardNew,
	}
}

func mustConnectClient() *broker.Client {
	cfg := broker.Config{
		URL:           "http://localhost:4223",
		ClientName:    "test-client",
		MaxReconnects: 3,
		ReconnectWait: time.Second,
	}

	client, appErr := broker.Connect(context.Background(), cfg)
	if appErr != nil {
		panic(appErr)
	}

	return client
}
