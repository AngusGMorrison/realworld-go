#!/bin/bash

set -eou pipefail

echo "Generating development data mount fixtures..."

echo "Checking for existence of host data directory '${REALWORLD_HOST_DATA_DIR}'..."
if [ -d "${REALWORLD_HOST_DATA_DIR}" ]; then
  echo "Host data directory already exists. Skipping generation."
  echo
  exit 0
fi

echo "Creating gitignored host data directory..."
mkdir "${REALWORLD_HOST_DATA_DIR}"
echo "Host data directory created."

echo "Generating development-only JWT RSA private key..."
ssh-keygen -t rsa -b 4096 -m PEM -f "${REALWORLD_HOST_DATA_DIR}/${REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME}" -N ""
echo "Development JWT key generated."

echo "Data mount fixtures generated successfully at '${REALWORLD_HOST_DATA_DIR}.'"
