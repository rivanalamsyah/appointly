# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION:-dev}" \
    -o /bin/api ./cmd/api

# Build Worker binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION:-dev}" \
    -o /bin/worker ./cmd/worker

# ---- Final Image ----
FROM gcr.io/distroless/static-debian12 AS final

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

ARG TARGET=api
COPY --from=builder /bin/${TARGET} /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
