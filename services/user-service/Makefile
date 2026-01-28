# services/user-service/Makefile
DB_URL=postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable
MIGRATION_DIR=internal/db/migrations

.PHONY: setup migrate sqlc build test

# Runs migrations and generates sqlc code
setup: migrate sqlc

# Pushes your auth schema to local Postgres
migrate:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" up

# Generates Go code into internal/db/generated
sqlc:
	sqlc generate

# Builds the service binary
build:
	go build -o ../../bin/user-service ./cmd/main.go

# Runs only user-service tests
test:
	go test -v ./...
