SHELL := /bin/sh

.DEFAULT_GOAL := help

.PHONY: help setup setup-proto dev

help:
	@printf "Relay development workflow commands.\n\n"
	@printf "Use these targets to prepare the workspace, install tooling, and start the local stack.\n\n"
	@printf "Available targets:\n"
	@printf "  make setup           Prepare the Relay workspace for local development\n"
	@printf "  make setup-proto     Install the Buf CLI for protobuf workflows\n"
	@printf "  make dev             Start the local development stack with Docker Compose\n"

setup:
	go run ./cli setup

setup-proto:
	npm install -g @bufbuild/buf

dev:
	go run ./cli dev
