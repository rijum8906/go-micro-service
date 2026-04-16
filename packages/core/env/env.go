// Package env provides environment configuration
package env

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Config struct {
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
	AppName string `env:"APP_NAME,required"`
	Port    int    `env:"PORT" envDefault:"8080"`

	// Services' Address
	UserServiceAddr    string `env:"USER_SERVICE_ADDR" envDefault:"user-service:8081"`
	GraphQLGatewayAddr string `env:"GRAPHQL_GATEWAY_ADDR" envDefault:"graphql-gateway:8080"`

	// NATS
	NATSURL        string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	NATSClientName string `env:"NATS_CLIENT_NAME" envDefault:"user-service"`

	// Postgres
	DBHost     string `env:"DB_HOST,required"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`

	// Redis
	RedisHost string `env:"REDIS_HOST,required"`
	RedisPort int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPass string `env:"REDIS_PASS"`

	// SMTP
	SMTPHost        string `env:"SMTP_HOST,required"`
	SMTPPort        int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername    string `env:"SMTP_USERNAME,required"`
	SMTPPassword    string `env:"SMTP_PASSWORD,required"`
	SMTPFromEmail   string `env:"SMTP_FROM_EMAIL,required"`
	SMTPFromName    string `env:"SMTP_FROM_NAME,required"`
	SMTPUseTLS      bool   `env:"SMTP_USE_TLS" envDefault:"true"`
	SMTPUseStartTLS bool   `env:"SMTP_USE_STARTTLS" envDefault:"false"`
	SMTPTimeout     int    `env:"SMTP_TIMEOUT" envDefault:"10"`
	SMTPRetries     int    `env:"SMTP_RETRIES" envDefault:"3"`

	// JWT
	JWTSecret    string `env:"JWT_SECRET,required"`
	ScopedSecret string `env:"SCOPED_SECRET,required"` // For scoped tokens

	// Logger
	LogLevel     string `env:"LOG_LEVEL" envDefault:"debug"`
	EnableJSON   bool   `env:"ENABLE_JSON" envDefault:"false"`
	EnableCaller bool   `env:"ENABLE_CALLER" envDefault:"false"`
	EnableStack  bool   `env:"ENABLE_STACK" envDefault:"false"`
	LogFile      string `env:"LOG_FILE" envDefault:"-"`

	// TTL
	SessionTTL      time.Duration `env:"SESSION_TTL" envDefault:"15m"`
	ScopedTokenTTL  time.Duration `env:"SCOPED_TOKEN_TTL" envDefault:"10m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"600h"`

	// CORS Policies
	CorsAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`
	CorsAllowedMethods []string `env:"CORS_ALLOWED_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	CorsAllowedHeaders []string `env:"CORS_ALLOWED_HEADERS" envDefault:"Content-Type, Authorization"`
}

// Load reads .env files (for local dev) and parses environment variables
func Load() (*Config, *apperror.AppError) {
	// 1. Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{}

	// 2. Parse variables into the struct
	if err := env.Parse(cfg); err != nil {
		return nil, apperror.ErrInternal.WithMessage(fmt.Sprintf("Failed to parse environment variables: %v", err))
	}

	return cfg, nil
}

// MustLoad is a helper for main.go that panics if config is missing
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Critical Config Error: %v", err)
	}
	return cfg
}
