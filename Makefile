SHELL := /bin/sh

.DEFAULT_GOAL := help

.PHONY: help setup setup-proto

help:
	@printf "Available targets:\n"
	@printf "  make setup SERVICE=<service>  Run a service setup target\n"
	@printf "  make setup-proto              Install Buf CLI\n"

setup:
	@if [ -z "$(SERVICE)" ]; then \
		printf "SERVICE is required. Usage: make setup SERVICE=<service>\n" >&2; \
		exit 1; \
	fi
	@$(MAKE) -C services/$(SERVICE) setup

setup-proto:
	npm install -g @bufbuild/buf
