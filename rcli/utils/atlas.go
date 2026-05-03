package utils

import (
	"fmt"
)

func GetDBURL(cfg *Environment) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		"localhost",
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func GetTestDBURL(cfg *Environment) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		"test_user",
		"test_password",
		"localhost",
		5433,
		GetTestDBName(cfg.AppName),
		cfg.DBSSLMode,
	)
}

func GetDynamicDBURL(useTestDB bool, cfg *Environment) string {
	if useTestDB {
		return GetTestDBURL(cfg)
	}
	return GetDBURL(cfg)
}

func GetDevDBURL(cfg *Environment) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		"localhost",
		cfg.DBPort,
		DevDBName,
		cfg.DBSSLMode,
	)
}

func GetTestDevURL(cfg *Environment) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		"test_user",
		"test_password",
		"localhost",
		5433,
		"test_db",
		cfg.DBSSLMode,
	)
}

func GetDynamicDevDBURL(useTestDB bool, cfg *Environment) string {
	if useTestDB {
		return GetTestDevURL(cfg)
	}
	return GetDevDBURL(cfg)
}

func GetMigrationDir() string {
	return "file://db/migrations"
}

func GetSchemaDir() string {
	return "file://db/schema.sql"
}
