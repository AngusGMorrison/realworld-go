# Load tasks.
-include tasks/Makefile.*

default: help

.PHONY: help
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

.PHONY: clean
## Remove all Make-generated artifacts.
clean: docker/clean generate/clean

.PHONY: build
## Compile the application.
build:
	CGO_ENABLED=1 GOFLAGS=-buildvcs=false go build -o ./bin/ ./cmd/server ./cmd/healthcheck

.PHONY: vulncheck
## Check dependencies for vulnerabilities.
vulncheck:
	govulncheck ./...
