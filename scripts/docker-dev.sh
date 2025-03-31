#!/bin/bash

# Exit on error
set -e

# Function to display usage
function display_usage {
  echo "Conveyor Development Environment"
  echo ""
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  up          Start development containers"
  echo "  down        Stop development containers"
  echo "  logs        View logs from all containers"
  echo "  rebuild     Rebuild and restart containers"
  echo "  frontend    View frontend logs only"
  echo "  backend     View backend logs only"
  echo "  redis       View redis logs only"
  echo ""
  echo "Options:"
  echo "  -h, --help  Display this help message"
}

# Check for help flag
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
  display_usage
  exit 0
fi

# Default command if none provided
COMMAND=${1:-up}

# Create necessary directories
mkdir -p data plugins tmp

# We'll create empty files instead of using Go commands
# since Go might not be installed on the user's machine
if [ ! -f "go.mod" ]; then
  echo "Creating empty go.mod file..."
  echo "module github.com/chip/conveyor" > go.mod
  echo "" >> go.mod
  echo "go 1.21" >> go.mod
fi

if [ ! -f "go.sum" ]; then
  echo "Creating empty go.sum file..."
  touch go.sum
fi

case $COMMAND in
  up)
    echo "Starting development environment..."
    docker-compose -f docker-compose.dev.yml up -d
    echo "Development environment started."
    echo "  Frontend: http://localhost:3000"
    echo "  Backend API: http://localhost:8080/api"
    ;;
  down)
    echo "Stopping development environment..."
    docker-compose -f docker-compose.dev.yml down
    echo "Development environment stopped."
    ;;
  logs)
    docker-compose -f docker-compose.dev.yml logs -f
    ;;
  rebuild)
    echo "Rebuilding development environment..."
    docker-compose -f docker-compose.dev.yml down
    docker-compose -f docker-compose.dev.yml build
    docker-compose -f docker-compose.dev.yml up -d
    echo "Development environment rebuilt and started."
    ;;
  frontend)
    docker-compose -f docker-compose.dev.yml logs -f frontend-dev
    ;;
  backend)
    docker-compose -f docker-compose.dev.yml logs -f backend-dev
    ;;
  redis)
    docker-compose -f docker-compose.dev.yml logs -f redis
    ;;
  *)
    echo "Error: Unknown command '$COMMAND'"
    display_usage
    exit 1
    ;;
esac 