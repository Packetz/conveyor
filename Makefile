.PHONY: build test lint clean dev docker-build docker-up docker-down

# Build the application
build:
	go build -o conveyor ./cli

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f conveyor
	rm -rf data
	rm -rf ui/dist

# Start development environment
dev:
	./scripts/dev.sh

# Build Docker image
docker-build:
	docker-compose build

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# Install dependencies
deps:
	go mod download
	go mod tidy
	cd ui && npm install

# Generate API documentation
docs:
	swag init -g api/server.go

# Run security scan
security-scan:
	gosec ./...
	trivy fs .

# Run all checks
check: lint test security-scan 