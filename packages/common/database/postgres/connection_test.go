package postgres_test

import (
	"context"
	"testing"

	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name         string
		cfg          postgres.Config
		wantErr      bool
		expectedCode int
	}{
		{
			name: "Invalid DSN - Connection Failure",
			cfg: postgres.Config{
				User:     "wrong_user",
				Password: "wrong_password",
				Host:     "localhost",
				Port:     5432,
				Database: "postgres",
				SSLMode:  "disable",
			},
			wantErr:      false,
			expectedCode: 500,
		},
		{
			name: "Valid DSN - Connection Success",
			cfg: postgres.Config{
				User:     "postgres",
				Password: "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "postgres",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPool, gotErr := postgres.Connect(context.Background(), tt.cfg)

			if tt.wantErr {
				if gotErr == nil {
					t.Fatal("Connect() expected an error but got nil")
				}
				// Verify it's using your custom AppError correctly
				if gotErr.StatusCode != tt.expectedCode {
					t.Errorf("Connect() error code = %v, want %v", gotErr.StatusCode, tt.expectedCode)
				}
				if gotPool != nil {
					t.Error("Connect() returned a pool on failure, expected nil")
				}
			} else {
				if gotErr != nil {
					t.Errorf("Connect() unexpected error: %v", gotErr)
				}
				if gotPool == nil {
					t.Error("Connect() returned nil pool on success")
				}
			}
		})
	}
}
