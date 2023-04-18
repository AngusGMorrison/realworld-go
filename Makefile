.PHONY: docker_build docker_create_volume docker_run docker_run_it

docker_build:
	docker build \
	--tag ${REALWORLD_IMAGE_NAME} \
	--build-arg volume_mount_path=${REALWORLD_VOLUME_MOUNT_PATH} \
	--build-arg port=${REALWORLD_PORT} \
	.

docker_create_volume:
	docker volume create ${REALWORLD_VOLUME_NAME}

docker_run: docker_build docker_create_volume
	docker run \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	--mount source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	realworld

docker_run_it: docker_build docker_create_volume
	docker run \
	--name=${REALWORLD_CONTAINER_NAME} \
	--interactive \
	--tty \
	--rm \
	--publish ${REALWORLD_PORT}:${REALWORLD_PORT} \
	--mount source=${REALWORLD_VOLUME_NAME},destination=${REALWORLD_VOLUME_MOUNT_PATH} \
	realworld
