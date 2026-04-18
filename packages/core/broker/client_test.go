package broker_test

import (
	"context"
	"testing"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg     broker.Config
		wantErr bool
		errCode apperror.ErrorCode
	}{
		{
			name:    "invalid config",
			cfg:     broker.Config{},
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name: "missing url",
			cfg: broker.Config{
				URL:           "",
				ClientName:    "test-client",
				MaxReconnects: 3,
				ReconnectWait: time.Second,
			},
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name: "wrong url",
			cfg: broker.Config{
				URL:           "http://localhost:4222",
				ClientName:    "test-client",
				MaxReconnects: 3,
				ReconnectWait: time.Second,
			},
			wantErr: true,
			errCode: apperror.CodeThirdParty,
		},
		{
			name: "valid config",
			cfg: broker.Config{
				URL:           "http://localhost:4223",
				ClientName:    "test-client",
				MaxReconnects: 3,
				ReconnectWait: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appErr := broker.Connect(context.Background(), tt.cfg)
			if tt.wantErr {
				if appErr == nil {
					t.Fatalf("Connect() = nil, want error")
				}

				if appErr.Code != tt.errCode {
					t.Fatalf("Connect() error code = %s, want %s", appErr.Code, tt.errCode)
				}
			} else {
				if appErr != nil {
					t.Fatalf("Connect() = %v, want nil", appErr)
				}
			}
		})
	}
}

func TestClient_IsConnected(t *testing.T) {
	cfg := broker.Config{
		URL:           "http://localhost:4223",
		ClientName:    "test-client",
		MaxReconnects: 3,
		ReconnectWait: time.Second,
	}

	client, appErr := broker.Connect(context.Background(), cfg)
	if appErr != nil {
		t.Fatalf("Connect() = %v, want nil", appErr)
	}

	isConnected := client.IsConnected()
	if !isConnected {
		t.Fatalf("IsConnected() = %v, want true", isConnected)
	}
}

func TestClient_Drain(t *testing.T) {
	cfg := broker.Config{
		URL:           "http://localhost:4223",
		ClientName:    "test-client",
		MaxReconnects: 3,
		ReconnectWait: time.Second,
	}

	// Connect
	client, appErr := broker.Connect(context.Background(), cfg)
	if appErr != nil {
		t.Fatalf("Connect() = %v, want nil", appErr)
	}

	// Drain
	appErr = client.Drain()
	if appErr != nil {
		t.Fatalf("Drain() = %v, want nil", appErr)
	}

	// Check Draining Status
	isDraining := client.Conn.IsDraining()
	if !isDraining {
		t.Fatalf("IsDraining() = %v, want true", isDraining)
	}
}

func TestClient_Close(t *testing.T) {
	cfg := broker.Config{
		URL:           "http://localhost:4223",
		ClientName:    "test-client",
		MaxReconnects: 3,
		ReconnectWait: time.Second,
	}

	// Connect
	client, appErr := broker.Connect(context.Background(), cfg)
	if appErr != nil {
		t.Fatalf("Connect() = %v, want nil", appErr)
	}

	// Close
	client.Close()

	// Check Conn Status
	isConnected := client.IsConnected()
	if isConnected {
		t.Fatalf("IsConnected() = %v, want false", isConnected)
	}
}
