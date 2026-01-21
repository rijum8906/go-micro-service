// Package redis
package redis

import (
	"fmt"

	goRedis "github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database int
}

func Connect(cfg Config) *goRedis.Client {
	client := goRedis.NewClient(&goRedis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})
	return client
}
