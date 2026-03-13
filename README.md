# Conveyor

**Conveyor** is a modern, extensible CI/CD platform designed for developer productivity and security. It features a pipeline execution engine, built-in security scanning, and a plugin architecture — all backed by a Go backend (Gin) with a React/TypeScript frontend (Vite + Material-UI).

## Key Features

- **Pipeline Engine**: Dynamic parallel execution with dependency management, conditional stages, and intelligent caching
- **YAML Pipeline Definitions**: Define pipelines in YAML with stage dependencies (`needs`), conditional execution (`when`), retry policies, notifications, and artifact management
- **Security Scanning**: Built-in security plugin for secret detection, vulnerability scanning, static analysis, license compliance, and SBOM generation
- **Plugin Architecture**: Extensible plugin system — plugins provide a manifest (capabilities, config schema, step types) and an execution function
- **Real-time Updates**: WebSocket support (`/ws`) streams pipeline events to the frontend as JSON
- **REST API**: Full API under `/api` for pipelines, jobs, security, plugins, and system health
- **Modern UI**: React/TypeScript frontend with Material-UI dark theme and persistent drawer navigation

## Getting Started

### Prerequisites

- Go 1.16+
- Node.js 18+
- Docker and Docker Compose (recommended for development)
- Redis (required for local development without Docker)

### Quick Start (Docker)

```bash
git clone https://github.com/Packetz/conveyor.git
cd conveyor

# Start frontend (3000), backend (8080), and Redis (6379)
./scripts/docker-dev.sh up
```

Access the UI at http://localhost:3000.

### Development

**Docker (recommended)** — includes hot reloading for both frontend and backend:

```bash
./scripts/docker-dev.sh up        # Start all services
./scripts/docker-dev.sh logs      # View logs
./scripts/docker-dev.sh rebuild   # Rebuild and restart
./scripts/docker-dev.sh down      # Stop services
```

**Local** — requires Go and Redis running locally:

```bash
make deps    # Install Go and npm dependencies
make dev     # Start the backend
cd ui && npm run dev   # Start the frontend (separate terminal)
```

**Frontend standalone:**

```bash
cd ui && npm install && npm run dev   # Vite dev server on :3000
```

### Build & Test

```bash
make build              # Build Go binary (output: ./conveyor)
make test               # go test -v ./...
make lint               # golangci-lint run (requires golangci-lint)
make check              # Run all checks (lint, test, security-scan)
make security-scan      # gosec + trivy (requires gosec and trivy)
make docs               # Generate API docs (requires swag)
make docker-build       # Build production Docker images
make clean              # Remove binary, data/, and ui/dist/
```

## Architecture

```
cli/main.go           — Entry point: initializes engine, registers plugins, starts API server
core/pipeline.go      — Pipeline engine (PipelineEngine): manages pipelines, jobs, plugins
core/loader/          — YAML pipeline loader: parses, validates, converts, and registers pipelines
api/server.go         — Gin HTTP server with WebSocket support and graceful shutdown
api/routes/           — Route handlers: pipeline.go, job.go, plugin.go, security.go, system.go
plugins/              — Plugin manager + built-in security scanning plugin
ui/                   — React/TypeScript frontend (Vite + Material-UI)
pipelines/            — YAML pipeline definitions loaded at startup
```

**Key patterns:**
- The `PipelineEngine` emits events through channels; the WebSocket endpoint (`/ws`) streams them to the frontend
- Plugins implement `Execute()` and `GetManifest()` — the security plugin demonstrates the full pattern
- At startup, `core/loader` scans `pipelines/` for `.yaml`/`.yml` files, validates and registers them with the engine

## Pipeline Configuration

Pipelines are defined in YAML with support for stage dependencies, conditional execution, security scanning plugins, notifications, and artifact management:

```yaml
name: secure-build
description: A secure CI/CD pipeline with security scanning

triggers:
  - type: push
    branches: [main, develop]
  - type: pull_request
    events: [opened, synchronize]

stages:
  - name: pre-build
    steps:
      - name: dependencies
        run: |
          npm ci
          go mod download

  - name: security-checks
    steps:
      - name: secret-scan
        plugin: security-scanner
        config:
          scanTypes: [secret]
          severityThreshold: HIGH
          failOnViolation: true

  - name: build
    needs: [pre-build, security-checks]
    steps:
      - name: build-backend
        run: go build -o bin/server ./cmd/server

  - name: deploy
    needs: [build]
    when:
      branch: main
    steps:
      - name: deploy-app
        run: echo "Deploying to production..."

notifications:
  - type: slack
    channel: "#builds"
    events: [success, failure]
```

## API Endpoints

All REST endpoints under `/api`:

| Endpoint | Description |
|----------|-------------|
| `GET/POST /api/pipelines` | List and create pipelines |
| `POST /api/pipelines/:id/execute` | Execute a pipeline |
| `POST /api/pipelines/import` | Import pipeline from YAML |
| `GET /api/pipelines/:id/jobs` | List jobs for a pipeline |
| `POST /api/pipelines/:id/jobs/:jobID/retry` | Retry a job |
| `GET/PUT /api/security/config` | Security configuration |
| `GET /api/security/scans` | Security scan results |
| `GET /api/plugins` | Plugin management |
| `GET /api/system/health` | Health check |
| `GET /api/system/metrics` | System metrics |
| `WS /ws` | Real-time event streaming |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
