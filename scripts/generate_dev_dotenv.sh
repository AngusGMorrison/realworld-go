#!/bin/bash

# This script generates the dev.env file required to build and run
# the application in development and CI. This allows us to dynamically
# define variables based on the host environment, such as the
# current working directory and the location of the Go compiler
# cache.

set -eou pipefail

GOCACHE=$(go env GOCACHE)
WORKDIR=$(pwd)

echo "Generating dev.env..."

cat > dev.env <<EOF
# Environment variables used in the build and run tasks in development and test.
# Overridden in production.
# MUST NOT contain secrets.

############
## Common ##
############

REALWORLD_APP_NAME=realworld

# Absolute path to the host data directory to be mounted into the app container.
REALWORLD_HOST_DATA_DIR=$WORKDIR/data

# Absolute path to the container directory where the host data directory will be
# mounted.
REALWORLD_DATA_MOUNT=/data

REALWORLD_PORT=8090

# Non-root user for all containers. Corresponds to the Distroless nonroot user.
REALWORLD_USER=65532

###########
## Build ##
###########

GOCACHE=$GOCACHE
REALWORLD_CONTAINER_NAME=realworld
REALWORLD_IMAGE_NAME=realworld

#########
## Run ##
#########

REALWORLD_DB_BASENAME=dev.db
REALWORLD_ENABLE_STACK_TRACE=true
REALWORLD_JWT_ISSUER="https://realworld.io"
REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME=jwtRS256.key
EOF

echo "dev.env generated!"
printf "Source and export it with\n\tset -a; source dev.env; set +a\n"