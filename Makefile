# Load tasks.
-include tasks/Makefile.*

# Load environment variables.
-include env/*.env
export

.PHONY: help test build

default: help

## Display this help message.
help:
	@printf "Available targets:\n\n"
		@awk '/^[a-zA-Z\-\_0-9%:\\]+/ { \
			helpMessage = match(lastLine, /^## (.*)/); \
			if (helpMessage) { \
				helpCommand = $$1; \
				helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
				gsub("\\\\", "", helpCommand); \
				gsub(":+$$", "", helpCommand); \
				printf "  \x1b[32;01m%-35s\x1b[0m %s\n", helpCommand, helpMessage; \
			} \
		} \
		{ lastLine = $$0 }' $(MAKEFILE_LIST) | sort -u
		@printf "\n"

## Run tests locally.
test: db/sqlc
	go test -race ./...

## Compile the application. CGO is required by the SQLite driver.
build:
	CGO_ENABLED=1 GOFLAGS=-buildvcs=false go build -o ./bin/server ./cmd/server
