# ==============================================================================
# Appointly Backend Production Dockerfile (Multi-Stage)
# ==============================================================================

# --- Stage 1: Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Install build dependencies, CA certificates, and timezones
RUN apk add --no-cache ca-certificates tzdata git

# Copy dependency manifests first for optimal layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and database migrations
COPY . .

# Build lightweight static binary
ARG VERSION=1.0.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /bin/appointly-api ./cmd/api

# --- Stage 2: Minimal Production Runtime ---
FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root system user & group
RUN addgroup -g 10001 -S appointly && \
    adduser -u 10001 -S appointly -G appointly

WORKDIR /app

# Copy binary, certificates, and migrations
COPY --from=builder /bin/appointly-api /app/appointly-api
COPY --from=builder /src/db/migrations /app/db/migrations

# Set permissions
RUN chown -R appointly:appointly /app

# Run as non-root user
USER appointly:appointly

EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/appointly-api"]
