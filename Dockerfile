# Dockerfile for Mockzure

# Development stage
FROM golang:1.23-alpine AS development

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache git build-base

# Copy go mod files (if exists)
COPY go.* ./
RUN go mod download 2>/dev/null || true

# Copy source
COPY . .

# Create data and logs directories
RUN mkdir -p /app/data /app/logs

# Expose port
EXPOSE 8090

# Build and run (for development with volume mounts)
CMD ["sh", "-c", "go build -o /tmp/mockzure main.go && /tmp/mockzure"]

# Production stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy source
COPY . .

# Build
RUN go build -o mockzure main.go

# Runtime image
FROM alpine:latest AS production

WORKDIR /app

# Copy binary
COPY --from=builder /app/mockzure .

# Copy config (will be overridden by volume mount in dev)
COPY --from=builder /app/config.json.example config.json

# Create data directory
RUN mkdir -p /app/data /app/logs

# Expose port
EXPOSE 8090

# Run
CMD ["./mockzure"]
