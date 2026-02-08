// Package redis contains functions for connecting to a redis database
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database int
}

// Connect creates a new Redis client and verifies connectivity
func Connect(ctx context.Context, cfg Config) (*redis.Client, *errors.AppError) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.User,
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	// Use Ping to ensure the connection is actually established
	if err := client.Ping(ctx).Err(); err != nil {
		// Wrap the internal Redis error into your custom AppError
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return client, nil
}
