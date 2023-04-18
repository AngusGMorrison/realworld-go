#!/bin/bash

# Load environment variables from .env file.
set -o allexport
source /app/.env
set +o allexport

# Start the server.
/app/server
