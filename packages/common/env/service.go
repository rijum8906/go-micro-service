// Package env
package env

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rijum8906/relay/packages/common/errors"
)

const (
	// Log Level
	LogLevelDevelopment = "development"
	LogLevelDebug       = "debug"
	LogLevelProduction  = "production"
	LogLevelTest        = "test"

	// App Environment
	AppEnvDevelopment = "development"
	AppEnvTest        = "test"
	AppEnvProduction  = "production"
)

type Env struct {
	// App Info
	AppName string
	AppEnv  string
	AppPort string

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSslMode  bool

	// Redis
	RedisHost     string
	RedisPort     int
	RedisUser     string
	RedisPassword string
	RedisDatabase int

	// JWT (Auth)
	JwtIssuer           string
	JwtSecret           string
	JwtExpiration       time.Duration
	ScopedJwtSecret     string
	ScopedJwtExpiration time.Duration

	// Email (SMTP) - Needed for Password Resets/Verification
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// Security / OTP
	OtpExpiration time.Duration
	BcryptCost    int

	// Storage (S3/Cloudflare R2) - For Profile Avatars
	StorageEndpoint  string
	StorageBucket    string
	StorageAccessKey string
	StorageSecretKey string
	StoragePublicKey string

	// CORS
	CorsAllowedOrigins []string
	CorsAllowedMethods []string
	CorsAllowedHeaders []string

	// Debug
	Debug bool

	// Logger
	LogLevel string
}

func Load() (*Env, *errors.AppError) {
	// Load .env file if it exists (useful for your Linux dev environment)
	err := godotenv.Load()
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	env := &Env{
		// App
		AppName: getString("APP_NAME", "user-service"),
		AppEnv:  getString("APP_ENV", "development"),
		AppPort: getString("APP_PORT", "8906"),

		// DB
		DBHost:     getString("DB_HOST", "localhost"),
		DBPort:     getInt("DB_PORT", 5432),
		DBUser:     getString("DB_USER", "postgres"),
		DBPassword: getString("DB_PASSWORD", "postgres"),
		DBName:     getString("DB_NAME", "postgres"),
		DBSslMode:  getBool("DB_SSL_MODE", false),

		// Redis
		RedisHost:     getString("REDIS_HOST", "localhost"),
		RedisPort:     getInt("REDIS_PORT", 6379),
		RedisUser:     getString("REDIS_USER", ""),
		RedisPassword: getString("REDIS_PASSWORD", ""),
		RedisDatabase: getInt("REDIS_DB", 0),

		// JWT
		JwtIssuer:           getString("JWT_ISSUER", "user-service"),
		JwtSecret:           getRequiredString("JWT_SECRET"), // Crashes if missing
		JwtExpiration:       getDuration("JWT_EXPIRATION", 15*time.Minute),
		ScopedJwtSecret:     getRequiredString("SCOPED_JWT_SECRET"), // Crashes if missing
		ScopedJwtExpiration: getDuration("SCOPED_JWT_EXPIRATION", 15*time.Minute),

		// SMTP
		SMTPHost:     getString("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getInt("SMTP_PORT", 587),
		SMTPUser:     getString("SMTP_USER", ""),
		SMTPPassword: getString("SMTP_PASSWORD", ""),
		SMTPFrom:     getString("SMTP_FROM", "no-reply@yourdomain.com"),

		// Security
		OtpExpiration: getDuration("OTP_EXPIRATION", 10*time.Minute),
		BcryptCost:    getInt("BCRYPT_COST", 12),

		// Storage
		StorageEndpoint:  getString("STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageBucket:    getString("STORAGE_BUCKET", "avatars"),
		StorageAccessKey: getString("STORAGE_ACCESS_KEY", "admin"),
		StorageSecretKey: getString("STORAGE_SECRET_KEY", "password123"),
		StoragePublicKey: getString("STORAGE_PUBLIC_KEY", "http://localhost:9000"),

		// CORS
		CorsAllowedOrigins: strings.Split(getString("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		CorsAllowedMethods: strings.Split(getString("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE"), ","),
		CorsAllowedHeaders: strings.Split(getString("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,Content-Length"), ","),

		// Debug
		Debug: getBool("DEBUG", true),

		// Logger
		LogLevel: getString("LOG_LEVEL", "debug"),
	}

	return env, nil
}

// getRequiredString is for critical variables that MUST exist for the app to function
func getRequiredString(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return v
}

func getString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
