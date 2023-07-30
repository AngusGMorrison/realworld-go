#!/bin/bash

# Initialize dependencies for local development, then run the app with live
# reloading.
#
# We run as root locally, since generating the RSA keypair requires the user
# to be present in /etc/hosts, but REALWORLD_USER doesn't exist on the Air
# image. This is fine for local development. Production containers always run
# as REALWORLD_USER.

# Exit on first error or undefined variable.
set -eu

# Check data volume exists.
echo "Checking for volume ${REALWORLD_VOLUME_NAME}..."
[[ -d ${REALWORLD_DATA_DIR} ]]
echo "Volume ${REALWORLD_VOLUME_NAME} exists. Continuing..."

# Create a new, passwordless RS256 keypair for local development if one doesn't
# already exist.
echo "Checking volume for an existing RSA keypair..."
PRIVATE_KEY_PATH="${REALWORLD_DATA_DIR}/${REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME}"
if [[ ! -f $PRIVATE_KEY_PATH ]]; then
  echo "No RSA keypair found. Generating a new one..."
  ssh-keygen -t rsa -b 4096 -m PEM -q -N "" -f "$PRIVATE_KEY_PATH"
  echo "RSA keypair generated."
fi
echo "RSA keypair found. Continuing..."

# Launch air, using exec to replace the shell process, which otherwise traps
# signals from the host.
echo "Launching air..."
exec /go/bin/air