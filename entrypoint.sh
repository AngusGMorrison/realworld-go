#!/bin/bash

set -o allexport
source /app/.env
set +o allexport

ls -ld $REALWORLD_VOLUME_MOUNT_PATH

/app/server
