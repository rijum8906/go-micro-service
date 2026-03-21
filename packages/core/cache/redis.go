// Package cache provides Redis connection logic
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Config struct {
	Host        string
	Port        int
	Password    string
	DB          int // Redis default is 0
	RetryCounts int
}

// Connect initializes a Redis client with a health check
func Connect(ctx context.Context, cfg Config) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        10,
		MinIdleConns:    5,
		MaxRetries:      3,
		ConnMaxIdleTime: 5 * time.Minute,
	})

	// Retry Logic
	retryCounts := 5
	if cfg.RetryCounts > 0 {
		retryCounts = cfg.RetryCounts
	}

	var lastErr error
	for i := range retryCounts {
		select {
		case <-ctx.Done():
			client.Close()
			return nil, apperror.ErrInternal.WithMessage("Redis connection cancelled by context")
		default:
			// Ping check
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			_, err := client.Ping(pingCtx).Result()
			cancel()

			if err == nil {
				return client, nil // Success!
			}

			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	client.Close()
	return nil, apperror.ErrInternal.WithMessage(fmt.Sprintf("Redis unreachable: %v", lastErr))
}
