package testutils

import (
	"os"

	"github.com/rijum8906/relay/packages/core/env"
)

func NewTestEnv() *env.Config {
	os.Setenv("APP_NAME", "Test App")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("SCOPED_SECRET", "test_scoped_secret")
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "test_username")
	os.Setenv("SMTP_PASSWORD", "test_password")
	os.Setenv("SMTP_FROM_EMAIL", "test_from_email")
	os.Setenv("SMTP_FROM_NAME", "Test From Name")

	config, appErr := env.Load()
	if appErr != nil {
		panic(appErr)
	}
	if config == nil {
		panic("Environment config is nil")
	}

	return config
}
