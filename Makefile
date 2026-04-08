SHELL := /bin/sh

.DEFAULT_GOAL := help

.PHONY: help setup setup-proto

help:
	@printf "Available targets:\n"
	@printf "  make setup           Setup the project\n"
	@printf "  make setup-proto     Install Buf CLI\n"
	@printf "  dev                  Start the docker compose\n"

setup:
	go run cmd/main.go --setup

setup-proto:
	npm install -g @bufbuild/buf

dev:
	go run cmd/main.go --dev
