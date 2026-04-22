package utils

import "fmt"

func GetDBURL(cfg *Environment) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func GetDevDBURL(cfg *Environment) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		"dev_"+cfg.DBName,
		cfg.DBSSLMode,
	)
}

func GetMigrationDir() string {
	return "file://db/migrations"
}

func GetSchemaDir() string {
	return "file://db/schema.sql"
}
