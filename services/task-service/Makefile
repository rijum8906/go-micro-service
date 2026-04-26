SHELL := /bin/sh

CLI_RUN := go run ./cmd/cli/main.go

.DEFAULT_GOAL := help

.PHONY: \
	help \
	run \
	setup \
	test \
	db-init \
	db-apply \
	db-migrate \
	db-status \
	db-new \
	db-schema \
	db-rehash \
	db-rollback \
	db-reset \
	test-setup \
	test-run \

help:
	@echo "Available targets:"
	@echo "  make run                        - Run the task service CLI"
	@echo "  make setup                      - Install development dependencies"
	@echo "  make test                       - Run task-service tests"
	@echo "  make db-init                    - Initialize database (create DBs)"
	@echo "  make db-apply                   - Apply pending Atlas migrations"
	@echo "  make db-migrate                 - Alias for db-apply"
	@echo "  make db-status                  - Show Atlas migration status"
	@echo "  make db-new NAME=xxx            - Generate an Atlas migration from schema changes"
	@echo "  make db-schema                  - Apply schema directly with Atlas"
	@echo "  make db-rehash                  - Recalculate atlas.sum"
	@echo "  make db-rollback COUNT=1        - Roll back Atlas migrations"
	@echo "  make db-reset                   - Clean database and reapply Atlas migrations"
	@echo "  make test-setup                 - Reset and migrate the test database"
	@echo "  make test-run                   - Run task-service tests"
	@echo ""
	@echo "Examples:"
	@echo "  make db-new NAME=add_users_table"
	@echo "  make db-rollback COUNT=2"
	@echo "  make test-setup"

run:
	@$(CLI_RUN) run

setup:
	@$(CLI_RUN) setup

test:
	@$(CLI_RUN) test run

db-apply:
	@$(CLI_RUN) db atlas apply

db-migrate: db-apply

db-status:
	@$(CLI_RUN) db atlas status

db-init:
	@$(CLI_RUN) db atlas init

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
	@$(CLI_RUN) test run

init-test: test-setup
