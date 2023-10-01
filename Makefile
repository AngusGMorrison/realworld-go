# Load tasks.
-include tasks/Makefile.*

.DEFAULT_GOAL := help

.PHONY: help
## Display this help message.
help:
	@scripts/print_make_help.sh $(shell realpath $(MAKEFILE_LIST))

.PHONY: build
## Build an optimized Docker image. Alias for docker/build.
build: docker/build

.PHONY: clean
## Remove all Make-generated artifacts.
clean: docker/clean generate/clean

.PHONY: generate
## Generate development dependencies.
generate: generate/queries generate/data_mount_fixtures

.PHONY: psql
## Run psql against the database. Alias for docker/exec/psql.
psql: docker/exec/psql

.PHONY: up
## Run the app interactively. Alias for docker/up.
up: docker/up

.PHONY: down
## Stop the app. Alias for docker/down.
down: docker/down

.PHONY: test
## Run the test suite. Alias for docker/test.
test: docker/test
