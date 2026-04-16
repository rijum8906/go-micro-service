SHELL := /bin/sh

CLI_RUN := go run ./cmd/cli/main.go

.DEFAULT_GOAL := help

.PHONY: \
	help \
	run \
	setup \
	test \
	db-apply \
	db-migrate \
	db-sql-apply \
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
	@echo "  make run                        - Run the notification service CLI stub"
	@echo "  make setup                      - Run setup CLI stub"
	@echo "  make test                       - Run test CLI stub"
	@echo "  make db-apply                   - Apply pending Atlas migrations"
	@echo "  make db-migrate                 - Alias for db-apply"
	@echo "  make db-sql-apply               - Apply embedded SQL migrations without Atlas"
	@echo "  make db-status                  - Show Atlas migration status"
	@echo "  make db-new NAME=xxx            - Generate an Atlas migration from schema changes"
	@echo "  make db-schema                  - Apply schema directly with Atlas"
	@echo "  make db-rehash                  - Recalculate atlas.sum"
	@echo "  make db-rollback COUNT=1        - Roll back Atlas migrations"
	@echo "  make db-reset                   - Clean database and reapply Atlas migrations"
	@echo "  make test-setup                 - Prepare the test environment"
	@echo "  make test-run                   - Run test command stub"
	@echo ""
	@echo "Examples:"
	@echo "  make db-new NAME=add_users_table"
	@echo "  make db-rollback COUNT=2"
	@echo "  make db-sql-apply"
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

db-sql-apply:
	@$(CLI_RUN) db sql apply

db-status:
	@$(CLI_RUN) db atlas status

db-init:
	@$(CLI_RUN) db atlas init

db-new:
	@$(CLI_RUN) db atlas new "$(NAME)"

db-schema:
	@$(CLI_RUN) db atlas schema

db-rehash:
	@$(CLI_RUN) db atlas rehash

db-rollback:
	@$(CLI_RUN) db atlas rollback $(or $(COUNT),1)

db-reset:
	@$(CLI_RUN) db atlas reset

test-setup:
	@$(CLI_RUN) test setup

test-run:
	@$(CLI_RUN) test run

init-test: test-start

