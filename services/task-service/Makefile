SHELL := /bin/sh

CLI_RUN := go run ./cmd/cli/main.go
GO_TEST := go test ./...

.DEFAULT_GOAL := help

.PHONY: \
	help \
	run \
	test \
	setup \

help:
	@echo "Available targets:"
	@echo "  make run                        - Run the task service"
	@echo "  make test                       - Run Go tests"
	@echo "  make setup                      - Run setup CLI stub"

run:
	@$(CLI_RUN) run

test:
	@$(CLI_RUN) test run

setup:
	@$(CLI_RUN) setup
