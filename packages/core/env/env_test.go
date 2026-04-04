package env_test

import (
	"os"
	"testing"

	"github.com/rijum8906/relay/packages/core/env"
)

func TestLoad(t *testing.T) {
	os.Setenv("APP_NAME", "Test App")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASS", "test_password")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("SCOPED_SECRET", "test_scoped_secret")

	config, appErr := env.Load()
	if appErr != nil {
		t.Errorf("Failed to load environment config: %v", appErr)
	}
	if config == nil {
		t.Errorf("Environment config is nil")
	}
}
