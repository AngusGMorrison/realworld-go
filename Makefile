include .env
export

.PHONY: test build gen_migration docker_build docker_create_volume docker_run docker_run_it docker_live

test:
	go test -race ./...

build: ## Compile the application. CGO is required by the SQLite driver.
	CGO_ENABLED=1 go build -o ./bin/server ./cmd/server

gen_migration: ## Generate a new, timestamped migration file.
	migrate create -ext sql -dir ./internal/repository/sqlite/migrations $(MIGRATION_NAME)

docker_build: ## Build the Docker image.
	docker build \
	--tag ${REALWORLD_IMAGE_NAME} \
	--build-arg volume_mount_path=${REALWORLD_VOLUME_MOUNT_PATH} \
	--build-arg port=${REALWORLD_PORT} \
	.

docker_create_volume: ## Create the persistence volume for the application.
	docker volume create ${REALWORLD_VOLUME_NAME}

docker_run: docker_create_volume ## Run the application in the background in a Docker container.
	docker run \
	--name=${REALWORLD_CONTAINER_NAME} \
	--env-file .env \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	--mount type=volume,source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	--rm \
	realworld

docker_run_it: docker_create_volume ## Run the application interactively in a Docker container.
	docker run \
	--interactive \
	--rm \
	--name=${REALWORLD_CONTAINER_NAME} \
	--env-file .env \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	--mount type=volume,source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	realworld

docker_live: ## Run the application interactively in a Docker container with live reloading.
	docker run \
	--interactive \
	--rm \
	--workdir ${REALWORLD_VOLUME_MOUNT_PATH} \
	--env air_wd=${REALWORLD_VOLUME_MOUNT_PATH} \
	--env GOFLAGS=-buildvcs=false \
	--env-file .env \
	--mount type=volume,source=.,destination=${REALWORLD_WORKDIR} \
	--mount source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	cosmtrek/air \
	-c .air.toml
