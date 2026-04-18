SHELL := /bin/sh

RUN_CMD := go run ./cmd/main.go
GO_TEST := go test ./...

.DEFAULT_GOAL := help

.PHONY: \
	help \
	run \
	test

help:
	@echo "Available targets:"
	@echo "  make run                        - Run the task service"
	@echo "  make test                       - Run Go tests"

run:
	@$(RUN_CMD)

test:
	@$(GO_TEST)
