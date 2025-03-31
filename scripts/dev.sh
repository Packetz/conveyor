#!/bin/bash

# Exit on error
set -e

# Load environment variables
if [ -f .env.development ]; then
    export $(cat .env.development | grep -v '^#' | xargs)
fi

# Create necessary directories
mkdir -p data plugins

# Check if Redis is running
if ! redis-cli ping &>/dev/null; then
    echo "Starting Redis..."
    docker-compose up -d redis
fi

# Start the development server
echo "Starting Conveyor development server..."
go run ./cli serve

# Trap SIGINT to clean up
trap 'echo "Shutting down..."; docker-compose down; exit 0' SIGINT SIGTERM 