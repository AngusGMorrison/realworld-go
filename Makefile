# Load tasks.
-include tasks/Makefile.*

.PHONY: help
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

.PHONY: build
## Build an optimized Docker image.
build: docker/build

.PHONY: clean
## Remove all Make-generated artifacts.
clean: docker/clean generate/clean

.PHONY: generate
## Generate development dependencies.
generate: generate/queries generate/data_mount_fixtures

.PHONY: run
## Run the app interactively.
run: docker/run

.PHONY: test
## Run the test suite.
test: docker/test
