package redis_test

import (
	"context"
	"testing"

	"github.com/rijum8906/relay/packages/common/database/redis"
	"github.com/rijum8906/relay/packages/common/errors"
)

func TestConnect(t *testing.T) {
	// 1. Start a local in-memory redis for testing

	tests := []struct {
		name         string
		cfg          redis.Config
		wantErr      bool
		expectedCode string
	}{
		{
			name: "Success - Correct Configuration",
			cfg: redis.Config{
				Host: "127.0.0.1",
				Port: 6379,
			},
			wantErr: false,
		},
		{
			name: "Failure - Connection Refused",
			cfg: redis.Config{
				Host: "127.0.0.1",
				Port: 1234,
			},
			wantErr:      true,
			expectedCode: errors.ErrDBConnection.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, appErr := redis.Connect(ctx, tt.cfg)

			if tt.wantErr {
				if appErr == nil {
					t.Errorf("Connect() expected error, got nil")
				} else if appErr.Code != tt.expectedCode {
					t.Errorf("Connect() status = %v, want %v", appErr.Code, tt.expectedCode)
				}
			} else {
				if appErr != nil {
					t.Errorf("Connect() unexpected error: %v", appErr)
				}
				if client == nil {
					t.Error("Connect() returned nil client on success")
				}
			}
		})
	}
}
