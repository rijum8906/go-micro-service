SHELL := /bin/sh

CLI_RUN := go run ./cli/

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
	test-start \
	test-run \
	test-stop \
	init-test \
	test-migrate \
	test-all \
	stop-test

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
	@echo "  make test-start                 - Start test dependencies"
	@echo "  make test-run                   - Run test command stub"
	@echo "  make test-stop                  - Stop test dependencies"
	@echo "  make init-test                  - Alias for test-start"
	@echo "  make test-migrate               - Run test migration stub"
	@echo "  make test-all                   - Run all test actions stub"
	@echo "  make stop-test                  - Alias for test-stop"
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

test-start:
	@$(CLI_RUN) test start

test-run:
	@$(CLI_RUN) test run

test-stop:
	@$(CLI_RUN) test stop

init-test: test-start

test-migrate:
	@$(CLI_RUN) test migrate

test-all:
	@$(CLI_RUN) test all

stop-test: test-stop
