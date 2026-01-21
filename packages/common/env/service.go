// Package env
package env

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	// App Info
	AppName string
	AppEnv  string

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

	// JWT
	JwtIssuer     string
	JwtSecret     string
	JwtExpiration time.Duration
}

func Load() (*Env, error) {
	_ = godotenv.Load() // ignore error in prod

	env := &Env{
		// App
		AppName: getString("APP_NAME", "user-service"),
		AppEnv:  getString("APP_ENV", "development"),

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
		JwtIssuer:     getString("JWT_ISSUER", "user-service"),
		JwtSecret:     getString("JWT_SECRET", "dev-secret"),
		JwtExpiration: getDuration("JWT_EXPIRATION", 15*time.Minute),
	}

	return env, nil
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
