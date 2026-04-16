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
	db-init \
	db-apply \
	db-sql-apply \
	db-status \
	db-new \
	db-schema \
	db-rehash \
	db-rollback \
	db-reset \
	test-setup \
	test-run

help:
	@echo "Available targets:"
	@echo "  make run                        - Run the user service"
	@echo "  make setup                      - Install sqlc and atlas (legacy)"
	@echo "  make test                       - Run Go tests"
	@echo ""
	@echo "Database:"
	@echo "  make db-init                    - Initialize database (create DBs)"
	@echo "  make db-apply                   - Apply pending Atlas migrations (Development)"
	@echo "  make db-sql-apply               - Apply embedded SQL migrations without Atlas (Local Use)"
	@echo "  make db-status                  - Show Atlas migration status"
	@echo "  make db-new NAME=xxx            - Generate an Atlas migration from schema changes"
	@echo "  make db-schema                  - Apply schema directly with Atlas"
	@echo "  make db-rehash                  - Recalculate atlas.sum"
	@echo "  make db-rollback COUNT=1        - Roll back Atlas migrations"
	@echo "  make db-reset                   - Clean database and reapply Atlas migrations"
	@echo ""
	@echo "Local test environment:"
	@echo "  make test-setup                 - Prepare the local test environment (runs migrations)"
	@echo "  make test-run                   - Run Go tests"
	@echo ""
	@echo "Examples:"
	@echo "  make db-new NAME=add_users_table"
	@echo "  make db-rollback COUNT=2"
	@echo "  make db-sql-apply"

run:
	@go run ./cmd/main.go

setup:
	@$(LEGACY_RUN) setup

test:
	@$(GO_TEST)

db-apply:
	@$(CLI_RUN) db atlas apply

db-sql-apply:
	@$(CLI_RUN) db sql apply

db-init:
	@$(CLI_RUN) db atlas init

db-status:
	@$(CLI_RUN) db atlas status

db-new:
	@$(CLI_RUN) db atlas new --name "$(NAME)"

db-schema:
	@$(CLI_RUN) db atlas schema

db-rehash:
	@$(CLI_RUN) db atlas rehash

db-rollback:
	@$(CLI_RUN) db atlas rollback --count $(or $(COUNT),1)

db-reset:
	@$(CLI_RUN) db atlas reset

test-setup:
	@$(CLI_RUN) test setup

test-run:
	@$(GO_TEST)
