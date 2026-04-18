package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/services/notification-service/app/config"
	"github.com/wneessen/go-mail"
)

func initDB(ctx context.Context, config *config.Env) (*pgxpool.Pool, *apperror.AppError) {
	pool, appErr := db.Connect(ctx, db.Config{
		Host:        config.DBHost,
		Port:        config.DBPort,
		User:        config.DBUser,
		Password:    config.DBPassword,
		DBName:      config.DBName,
		SSLMode:     config.DBSSLMode,
		RetryCounts: 5,
	})
	if appErr != nil {
		return nil, appErr
	}

	return pool, nil
}

func initCache(ctx context.Context, config *config.Env) (*redis.Client, *apperror.AppError) {
	cache, appErr := cache.Connect(ctx, cache.Config{
		Host:        config.RedisHost,
		Port:        config.RedisPort,
		DB:          config.RedisDB,
		Password:    config.RedisPass,
		RetryCounts: 5,
	})

	if appErr != nil {
		return nil, appErr
	}

	return cache, nil
}

func initNATS(ctx context.Context, config *config.Env) (*broker.Client, *apperror.AppError) {
	client, appErr := broker.Connect(ctx, broker.Config{
		URL:        config.NATSURL,
		ClientName: config.NATSClientName,
	})
	if appErr != nil {
		return nil, appErr
	}

	return client, nil
}

func initMailer(ctx context.Context, config *config.Env) (*mail.Client, *apperror.AppError) {
	mailer, appErr := mailer.Connect(getMailerConfig(config))

	if appErr != nil {
		return nil, appErr
	}

	return mailer, nil
}

func getMailerConfig(config *config.Env) mailer.Config {
	return mailer.Config{
		Host:        config.SMTPHost,
		Port:        config.SMTPPort,
		Username:    config.SMTPUsername,
		Password:    config.SMTPPassword,
		FromEmail:   config.SMTPFromEmail,
		FromName:    config.SMTPFromName,
		UseStartTLS: config.SMTPUseStartTLS,
		UseTLS:      config.SMTPUseTLS,
		Retries:     config.SMTPRetries,
		Timeout:     time.Minute,
	}
}
