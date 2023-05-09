.PHONY: test build docker_build docker_create_volume docker_run docker_run_it

test:
	go test -race ./...

build: ## Compile the application. CGO is required by the SQLite driver.
	CGO_ENABLED=1 go build -o ./bin/server ./cmd/server

docker_build: ## Build the Docker image.
	docker build \
	--tag ${REALWORLD_IMAGE_NAME} \
	--build-arg volume_mount_path=${REALWORLD_VOLUME_MOUNT_PATH} \
	--build-arg port=${REALWORLD_PORT} \
	.

docker_create_volume: ## Create the persistence volume for the application.
	docker volume create ${REALWORLD_VOLUME_NAME}

# Mandatory environment variables to be passed to the application via Docker.
docker_env_flags = \
	--env REALWORLD_JWT_RSA_PRIVATE_KEY_PEM \
	--env REALWORLD_VOLUME_MOUNT_PATH \
	--env REALWORLD_DB_BASENAME

docker_run: docker_create_volume ## Run the application in the background in a Docker container.
	docker run \
	--name=${REALWORLD_CONTAINER_NAME} \
	$(docker_env_flags) \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	--mount source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	realworld

docker_run_it: docker_create_volume ## Run the application interactively in a Docker container.
	docker run \
	--name=${REALWORLD_CONTAINER_NAME} \
	$(docker_env_flags) \
	--interactive \
	--tty \
	--rm \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	--mount source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	realworld
