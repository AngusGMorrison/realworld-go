# Load tasks.
-include tasks/Makefile.*

# Load environment variables.
-include env/*.env
export

default: help

## Display this help message.

.PHONY: help
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
.PHONY: test
test:
	go test -race -v ./...

## Compile the application. CGO is required by the SQLite driver.
.PHONY: build
build:
	CGO_ENABLED=1 GOFLAGS=-buildvcs=false go build -o ./bin/server ./cmd/server

.PHONY: vulncheck
vulncheck: generate
	govulncheck ./...
