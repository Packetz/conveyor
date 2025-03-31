# Conveyor

**Conveyor** is a modern, extensible CI/CD platform designed for developer productivity and security. It combines the best features of existing CI/CD tools while providing a streamlined experience through a modular plugin system.

![Conveyor CI/CD](https://via.placeholder.com/800x400?text=Conveyor+CI/CD)

## Key Features

- **Smart Pipeline Engine**: Dynamic parallel execution with dependency management and intelligent caching
- **Comprehensive Security Scanning**: Built-in security for code, dependencies, and configurations
- **Plugin Architecture**: Extend functionality with plugins for various integrations
- **Elegant UI**: Modern, responsive interface for monitoring and managing pipelines
- **REST API**: Full API for integration with other systems
- **Real-time Updates**: WebSocket support for live pipeline status updates

## Security Features

Conveyor includes a robust security scanning system that provides:

- Secret detection to prevent credential leaks
- Vulnerability scanning for dependencies
- Static code analysis for security issues
- License compliance checking
- SBOM (Software Bill of Materials) generation

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker and Docker Compose (for development)

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/Packetz/conveyor.git
   cd conveyor
   ```

2. Start the development environment:
   ```bash
   ./scripts/docker-dev.sh up
   ```

3. Access the UI at http://localhost:3000

### Development

For development, we use Docker Compose to set up a consistent environment:

```bash
# Start the development environment
./scripts/docker-dev.sh up

# Rebuild services
./scripts/docker-dev.sh rebuild

# View logs
./scripts/docker-dev.sh logs

# Shut down
./scripts/docker-dev.sh down
```

Both frontend and backend support hot reloading for rapid development.

## Architecture

Conveyor is built with a modular architecture:

- **Core**: Pipeline engine written in Go
- **API**: REST API and WebSocket support
- **Plugins**: Extensible plugin system
- **UI**: React/TypeScript frontend

## Pipeline Configuration

Pipelines are defined in YAML format:

```yaml
name: sample-pipeline
description: A sample pipeline

triggers:
  - type: push
    branches: [main, develop]

stages:
  - name: build
    steps:
      - name: checkout
        run: git clone https://github.com/user/repo.git

      - name: build
        run: go build -o app ./cmd/app

  - name: test
    steps:
      - name: unit-tests
        run: go test ./...

  - name: security
    steps:
      - name: security-scan
        plugin: security-scanner
        config:
          scanTypes: [secret, vulnerability, code]
          severityThreshold: HIGH
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 