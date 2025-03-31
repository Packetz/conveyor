# Build stage for Go backend
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the backend
RUN CGO_ENABLED=0 GOOS=linux go build -o conveyor ./cli

# Build stage for React frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY ui/package*.json ./
RUN npm install

# Copy source code
COPY ui/ .

# Build the frontend
RUN npm run build

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from backend builder
COPY --from=backend-builder /app/conveyor .

# Copy frontend build from frontend builder
COPY --from=frontend-builder /app/build ./ui/dist

# Copy configuration and plugins
COPY plugins ./plugins
COPY config ./config

# Create necessary directories
RUN mkdir -p /app/data

# Expose ports
EXPOSE 8080

# Set environment variables
ENV CONVEYOR_DATA_DIR=/app/data
ENV CONVEYOR_PLUGINS_DIR=/app/plugins

# Run the application
CMD ["./conveyor", "serve"] 