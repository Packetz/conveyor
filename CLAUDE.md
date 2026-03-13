# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Conveyor is a CI/CD platform with pipeline execution, security scanning, and a plugin architecture. Go backend (Gin) with React/TypeScript frontend (Vite + Material-UI).

## Prerequisites

Required external tools (not installed by `make deps`):
- **Go 1.16+** — backend compiler
- **golangci-lint** — `make lint`
- **gosec**, **trivy** — `make security-scan`
- **swag** — `make docs` (API doc generation)
- **air** — hot reload in Docker dev (included in dev container)
- **Node 18+** — frontend

## Build & Development Commands

```bash
# Build
make build              # Build Go binary (output: ./conveyor)
make docker-build       # Build production Docker images
make clean              # Remove binary, data/, and ui/dist/

# Test & Lint
make test               # go test -v ./...
make lint               # golangci-lint run
make check              # Run all checks (lint, test, security-scan)
make security-scan      # gosec + trivy
make docs               # Generate API docs (swag init)

# Development (Docker — recommended)
./scripts/docker-dev.sh up        # Start frontend (3000), backend (8080), Redis (6379)
./scripts/docker-dev.sh logs      # View logs
./scripts/docker-dev.sh rebuild   # Rebuild and restart
./scripts/docker-dev.sh down      # Stop services

# Development (local)
make dev                # Requires Go, Redis running locally

# Frontend (standalone)
cd ui && npm install && npm run dev   # Vite dev server on :3000

# Dependencies
make deps               # Install Go and npm dependencies
```

## Architecture

### Backend (Go)

- **`cli/main.go`** — Entry point. Initializes the pipeline engine, registers plugins, sets up sample data, and starts the API server.
- **`core/pipeline.go`** — Central pipeline engine (`PipelineEngine`). Manages pipelines, jobs, and plugins with RWMutex for thread safety. Event-driven via channels for real-time updates. Key types: `Pipeline`, `Stage`, `Step`, `Job`, `Event`.
- **`api/server.go`** — Gin HTTP server with WebSocket support (`/ws` endpoint for real-time event streaming). Graceful shutdown with context.
- **`api/routes/`** — Route handlers grouped by domain: `pipeline.go`, `job.go`, `plugin.go`, `security.go`, `system.go`.
- **`plugins/plugin.go`** — Plugin manager. Plugins implement `Execute()` and `GetManifest()`. Loaded from `manifest.json` + `.so` binary.
- **`plugins/security/`** — Security scanning plugin (secret scan, vulnerability scan, license check, code scan, SBOM generation). Configuration schema in `manifest.json`.
- **`core/loader/`** — YAML pipeline loader. Parses pipeline YAML files, validates structure, converts to core types, and loads from the `pipelines/` directory. Key files: `parse.go`, `validator.go`, `convert.go`, `slugify.go`, `loader.go`, `types.go`.
- **`pipelines/`** — Directory for pipeline YAML definitions loaded at startup (e.g., `secure-build.yaml`).

### Frontend (React/TypeScript)

- **`ui/src/App.tsx`** — Material-UI dark theme layout with persistent drawer navigation.
- **`ui/src/pages/`** — Dashboard, Pipelines, Plugins, Settings views.
- **`ui/src/services/api.ts`** — Axios client for backend API calls.
- **`ui/vite.config.ts`** — Dev server proxies `/api` and `/ws` to backend on port 8080.

### Key Patterns

- **Event system**: `PipelineEngine` emits events through channels; WebSocket endpoint streams them to the frontend as JSON.
- **Plugin interface**: All plugins provide a manifest (capabilities, config schema, step types) and an execution function. The security plugin demonstrates the full pattern.
- **Pipeline YAML**: Pipelines define stages with dependency ordering (`needs`), conditional execution (`when`), retry policies, and caching. See `samples/pipelines/secure-build.yaml` for a complete example.
- **YAML pipeline loader**: At startup, `core/loader` scans `pipelines/` for `.yaml`/`.yml` files, parses and validates them, converts to core types, and registers them with the engine. Pipelines can also be imported at runtime via the API.

### Infrastructure

- **Docker Compose dev** (`docker-compose.dev.yml`): frontend-dev (Node 18 + Vite hot reload), backend-dev (Go 1.21 + `air` hot reload, Delve on port 2345), Redis 7.
- **Docker Compose prod** (`docker-compose.yml`): Multi-stage build — Go binary + React static assets served from Alpine.

## API Structure

All REST endpoints under `/api`:
- `/api/pipelines` — CRUD + `/execute`, `/jobs`, `/jobs/:jobID/retry`, `/import` (POST, load from YAML)
- `/api/security` — `/config`, `/scans`
- `/api/plugins` — Plugin management
- `/api/system` — Health, metrics
- `/ws` — WebSocket real-time events
