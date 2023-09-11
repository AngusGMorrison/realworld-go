#!/bin/bash

# Initialize the data volume under the ownership of the user ID shared by all
# containers. Does nothing if the volume already exists.
#
# Without this initialization script, the volume is owned by the user of the
# first container to mount it, with unpredictable results for other containers.

# Exit on undefined variables or pipeline failures.
set -uo pipefail

# Returns 0 if the volume exists, 1 otherwise.
volume_exists() {
  docker volume ls -f name="${REALWORLD_VOLUME_NAME}" \
    | awk '{print $NF}' \
    | grep -q "^${REALWORLD_VOLUME_NAME}$"
}

create_volume() {
  docker volume create "${REALWORLD_VOLUME_NAME}"
}

# Creates the volume and initializes it with the non-root user ID shared by all
# containers. The choice of the temporary busybox container is arbitrary.
initialize_volume() {
  docker run \
    --rm \
    --workdir /tmp/data \
    --mount type=volume,src="${REALWORLD_VOLUME_NAME}",dst=/tmp/data \
    busybox \
    sh -c "touch .initialized && chmod -R 0760 . && chown -R ${REALWORLD_USER}:${REALWORLD_USER} . && rm .initialized"
}

remove_volume() {
  echo "Removing uninitialized volume ${REALWORLD_VOLUME_NAME}..."
  docker volume rm "${REALWORLD_VOLUME_NAME}"
}

echo "Checking for volume ${REALWORLD_VOLUME_NAME}..."
if volume_exists; then
  echo "Volume ${REALWORLD_VOLUME_NAME} already exists. Skipping creation."
  exit 0
fi
echo "Volume ${REALWORLD_VOLUME_NAME} does not exist."

echo "Creating volume ${REALWORLD_VOLUME_NAME}..."
if ! create_volume; then
  echo "Failed to create volume ${REALWORLD_VOLUME_NAME}."
  exit 1
fi
echo "Volume ${REALWORLD_VOLUME_NAME} created."

echo "Initializing volume ${REALWORLD_VOLUME_NAME}..."
if initialize_volume; then
  echo "Volume ${REALWORLD_VOLUME_NAME} initialized."
  exit 0
fi
echo "Failed to initialize volume ${REALWORLD_VOLUME_NAME}."

if remove_volume; then
  echo "Volume ${REALWORLD_VOLUME_NAME} removed."
  exit 1
fi

echo "Failed to remove volume ${REALWORLD_VOLUME_NAME}."
echo "Please remove it manually and try again."
exit 1
