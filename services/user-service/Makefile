SHELL := /bin/sh

CLI_RUN := go run ./cmd/cli/main.go
LEGACY_RUN := go run ./cmd/test-cli/main.go
GO_TEST := go test -v ./...

.DEFAULT_GOAL := help

.PHONY: \
	help \
	run \
	setup \
	test \
	db-apply \
	db-migrate \
	db-status \
	db-new \
	db-schema \
	db-rehash \
	db-rollback \
	db-reset \
	test-setup \
	test-start \
	test-run \
	test-stop \
	init-test \
	test-migrate \
	test-all \
	stop-test

help:
	@echo "Available targets:"
	@echo "  make run                        - Run the user service"
	@echo "  make setup                      - Install sqlc and atlas"
	@echo "  make test                       - Run Go tests"
	@echo ""
	@echo "Database:"
	@echo "  make db-apply                   - Apply pending migrations"
	@echo "  make db-migrate                 - Alias for db-apply"
	@echo "  make db-status                  - Show migration status"
	@echo "  make db-new NAME=xxx            - Generate migration from schema changes"
	@echo "  make db-schema                  - Apply schema directly"
	@echo "  make db-rehash                  - Recalculate atlas.sum"
	@echo "  make db-rollback COUNT=1        - Roll back migrations"
	@echo "  make db-reset                   - Clean database and reapply migrations"
	@echo ""
	@echo "Local test environment:"
	@echo "  make test-setup                 - Prepare the local test environment"
	@echo "  make test-start                 - Start local test containers"
	@echo "  make test-run                   - Run Go tests"
	@echo "  make test-stop                  - Stop local test containers"
	@echo "  make test-migrate               - Apply embedded SQL migrations to the test database"
	@echo "  make test-all                   - Start test env, migrate, and run tests"
	@echo ""
	@echo "Compatibility aliases:"
	@echo "  make init-test                  - Alias for test-start"
	@echo "  make stop-test                  - Alias for test-stop"
	@echo ""
	@echo "Examples:"
	@echo "  make db-new NAME=add_users_table"
	@echo "  make db-rollback COUNT=2"
	@echo "  make db-apply DB_NAME=mydb"

run:
	@go run ./cmd/main.go

setup:
	@$(LEGACY_RUN) setup

test:
	@$(GO_TEST)

db-apply:
	@$(CLI_RUN) db apply

db-migrate: db-apply

db-status:
	@$(CLI_RUN) db status

db-new:
	@$(CLI_RUN) db new --name "$(NAME)"

db-schema:
	@$(CLI_RUN) db schema

db-rehash:
	@$(CLI_RUN) db rehash

db-rollback:
	@$(CLI_RUN) db rollback --count $(or $(COUNT),1)

db-reset:
	@$(CLI_RUN) db reset

test-setup:
	@$(CLI_RUN) test setup

test-start:
	@$(CLI_RUN) test start

test-run:
	@$(GO_TEST)

test-stop:
	@$(CLI_RUN) test stop

init-test: test-start

test-migrate:
	@$(LEGACY_RUN) env-migrate

test-all:
	@$(LEGACY_RUN) env-test

stop-test: test-stop
