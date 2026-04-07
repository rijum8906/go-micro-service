SHELL := /bin/sh

# Database configuration
DB_NAME ?= postgres
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_HOST ?= localhost
DB_PORT ?= 5432

# Test database configuration
TEST_DB_CONTAINER_NAME := test_postgres_db
TEST_DB_PORT := 5433
TEST_DB_USER := test_user
TEST_DB_PASSWORD := test_password
TEST_DB_NAME := test_db
TEST_DB_SSL_MODE := disable
TEST_DB_URL := postgres://$(TEST_DB_USER):$(TEST_DB_PASSWORD)@localhost:$(TEST_DB_PORT)/$(TEST_DB_NAME)?sslmode=$(TEST_DB_SSL_MODE)&search_path=public


# Connection URLs
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable&search_path=public
DEV_DB_URL := docker://postgres/17/dev?search_path=public

.DEFAULT_GOAL := help

.PHONY: help setup db-new db-apply db-status db-schema

help:
	@echo "Available targets:"
	@echo "  make setup                 - Install tools (sqlc, atlas)"
	@echo "  make db-new NAME=xxx       - Generate migration from schema changes"
	@echo "  make db-apply              - Apply pending migrations"
	@echo "  make db-status             - Show migration status"
	@echo "  make db-schema             - Apply schema directly (initial setup)"
	@echo "  make db-init               - Initialize migration system"
	@echo "  make db-rehash             - Recalculate atlas.sum"
	@echo ""
	@echo "Examples:"
	@echo "  make db-new NAME=add_users_table"
	@echo "  make db-apply DB_NAME=mydb"

setup:
	@echo "Installing sqlc..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Installing atlas..."
	curl -sSf https://atlasgo.sh | sh
	@echo "Setup complete! Run 'make db-init' to initialize migrations."


db-new:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make db-new NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "Generating migration: $(NAME)"
	atlas migrate diff $(NAME) \
		--dir "file://db/migrations" \
		--to "file://db/schema.sql" \
		--dev-url "$(DEV_DB_URL)"
	@echo "Migration generated. Review the file in db/migrations/"

db-apply:
	@echo "Applying migrations to $(DB_NAME) database..."
	atlas migrate apply \
		--url "$(DB_URL)" \
		--dir "file://db/migrations"
	@echo "Migrations applied."

db-status:
	@echo "Migration status for $(DB_NAME):"
	atlas migrate status \
		--url "$(DB_URL)" \
		--dir "file://db/migrations"

db-schema:
	@echo "Applying schema directly to $(DB_NAME) database..."
	atlas schema apply \
		--url "$(DB_URL)" \
		--to "file://db/schema.sql" \
		--dev-url "$(DEV_DB_URL)" \
		--auto-approve
	@echo "Schema applied."

db-rehash:
	@echo "Rehashing migration directory..."
	atlas migrate hash --dir "file://db/migrations"
	@echo "atlas.sum updated."


db-rollback:
	@echo "Rolling back $(or $(COUNT),1) migration(s)..."
	atlas migrate down $(or $(COUNT),1) \
		--url "$(DB_URL)" \
		--dir "file://db/migrations" \
		--dev-url "$(DEV_DB_URL)" \

init-test:
	@echo "Initializing test"
	@echo "Starting PostgreSQL test container"
	docker run -d \
		--name $(TEST_DB_CONTAINER_NAME) \
		-e POSTGRES_USER=$(TEST_DB_USER) \
		-e POSTGRES_PASSWORD=$(TEST_DB_PASSWORD) \
		-e POSTGRES_DB=$(TEST_DB_NAME) \
		-p $(TEST_DB_PORT):5432 \
		postgres:16-alpine \
		-c 'log_statement=all'
	# Wait for PostgreSQL to be ready
	@echo "Waiting for PostgreSQL to be ready..."
	@for i in $$(seq 1 30); do \
		if docker exec $(TEST_DB_CONTAINER_NAME) pg_isready -U $(TEST_DB_USER) > /dev/null 2>&1; then \
			echo "PostgreSQL is ready!"; \
			break; \
		fi; \
		echo "Waiting... $$i/30"; \
		sleep 1; \
		if [ $$i -eq 30 ]; then \
			echo "Timeout waiting for PostgreSQL"; \
			exit 1; \
		fi; \
	done
	@echo "Applying migrations in database"
		atlas migrate apply \
		--url "$(TEST_DB_URL)" \
		--dir "file://db/migrations"
	@echo "Migrations applied."
	@echo "========================================="
	@echo "Test environment ready!"
	@echo "Database URL: $(TEST_DB_URL)"
	@echo "========================================="

test:
	go test -v ./...
