# =============================================================================
# Appointly — Makefile
# =============================================================================
.DEFAULT_GOAL := help
SHELL := /bin/bash

# --- Config ------------------------------------------------------------------
BACKEND_DIR   := backend
FRONTEND_DIR  := frontend
MIGRATION_DIR := $(BACKEND_DIR)/db/migrations
BINARY_DIR    := $(BACKEND_DIR)/bin
API_BINARY    := $(BINARY_DIR)/api
WORKER_BINARY := $(BINARY_DIR)/worker

# Load .env if present
ifneq (,$(wildcard .env))
  include .env
  export
endif

DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# --- Colors ------------------------------------------------------------------
CYAN  := \033[0;36m
GREEN := \033[0;32m
RESET := \033[0m

# =============================================================================
# HELP
# =============================================================================
.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "  $(CYAN)Appointly$(RESET) — Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-25s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# =============================================================================
# INFRASTRUCTURE
# =============================================================================
.PHONY: infra-up infra-down infra-reset infra-logs

infra-up: ## Start Docker infrastructure (postgres, redis, minio, mailhog)
	docker compose up -d postgres redis minio mailhog
	@echo "$(GREEN)Infrastructure started. Waiting for postgres...$(RESET)"
	@sleep 3
	@docker compose exec postgres pg_isready -U $(DB_USER) -d $(DB_NAME) || true

infra-down: ## Stop Docker infrastructure
	docker compose down

infra-reset: ## Stop, remove volumes, and restart infrastructure
	docker compose down -v
	$(MAKE) infra-up

infra-logs: ## Tail infrastructure logs
	docker compose logs -f postgres redis minio

# =============================================================================
# DATABASE MIGRATIONS
# =============================================================================
.PHONY: migrate-up migrate-down migrate-reset migrate-create migrate-status migrate-version

migrate-up: ## Run all pending migrations
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path $(MIGRATION_DIR) \
		-database "$(DB_URL)" \
		up
	@echo "$(GREEN)Migrations applied.$(RESET)"

migrate-down: ## Rollback the last migration
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path $(MIGRATION_DIR) \
		-database "$(DB_URL)" \
		down 1

migrate-reset: ## Rollback ALL migrations (DESTRUCTIVE)
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path $(MIGRATION_DIR) \
		-database "$(DB_URL)" \
		down -all

migrate-status: ## Show migration status
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path $(MIGRATION_DIR) \
		-database "$(DB_URL)" \
		version

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=create_users)
endif
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "$(GREEN)Migration files created in $(MIGRATION_DIR)$(RESET)"

# =============================================================================
# CODE GENERATION
# =============================================================================
.PHONY: sqlc-gen openapi-validate

sqlc-gen: ## Regenerate sqlc database queries
	cd $(BACKEND_DIR) && sqlc generate

openapi-validate: ## Validate OpenAPI spec
	npx -y @redocly/cli@latest lint $(BACKEND_DIR)/api/openapi.yaml

# =============================================================================
# DEVELOPMENT SERVERS
# =============================================================================
.PHONY: dev-api dev-worker dev-frontend dev-all

dev-api: ## Run API server with live reload (requires air)
	cd $(BACKEND_DIR) && air -c .air.toml

dev-worker: ## Run background worker with live reload
	cd $(BACKEND_DIR) && air -c .air.worker.toml

dev-frontend: ## Run Astro frontend dev server
	cd $(FRONTEND_DIR) && npm run dev

# =============================================================================
# TESTING
# =============================================================================
.PHONY: test-unit test-integration test-all test-coverage

test-unit: ## Run unit tests (no external deps required)
	cd $(BACKEND_DIR) && go test -v -short -count=1 ./...

test-integration: ## Run integration tests (requires infra-up)
	cd $(BACKEND_DIR) && go test -v -count=1 -tags=integration ./...

test-all: ## Run all tests
	$(MAKE) test-unit
	$(MAKE) test-integration

test-coverage: ## Generate test coverage report
	cd $(BACKEND_DIR) && go test -coverprofile=coverage.out ./...
	cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: $(BACKEND_DIR)/coverage.html$(RESET)"

# =============================================================================
# LINTING & FORMATTING
# =============================================================================
.PHONY: lint lint-go lint-frontend fmt fmt-go fmt-frontend

lint: lint-go lint-frontend ## Run all linters

lint-go: ## Run Go linter
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint-frontend: ## Run frontend linter
	cd $(FRONTEND_DIR) && npm run lint

fmt: fmt-go fmt-frontend ## Format all code

fmt-go: ## Format Go code
	cd $(BACKEND_DIR) && gofmt -w .
	cd $(BACKEND_DIR) && goimports -w .

fmt-frontend: ## Format frontend code
	cd $(FRONTEND_DIR) && npm run format

typecheck: ## Run TypeScript type check
	cd $(FRONTEND_DIR) && npm run typecheck

# =============================================================================
# BUILD
# =============================================================================
.PHONY: build-api build-worker build-frontend build-all

build-api: ## Build production API binary
	@mkdir -p $(BINARY_DIR)
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build \
		-ldflags="-s -w -X main.Version=$(APP_VERSION)" \
		-o ../$(API_BINARY) ./cmd/api

build-worker: ## Build production worker binary
	@mkdir -p $(BINARY_DIR)
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build \
		-ldflags="-s -w -X main.Version=$(APP_VERSION)" \
		-o ../$(WORKER_BINARY) ./cmd/worker

build-frontend: ## Build production frontend
	cd $(FRONTEND_DIR) && npm run build

build-all: build-api build-worker build-frontend ## Build everything

# =============================================================================
# DOCKER
# =============================================================================
.PHONY: docker-build docker-up docker-down

docker-build: ## Build all Docker images
	docker compose build

docker-up: ## Start full application stack
	docker compose up -d

docker-down: ## Stop full application stack
	docker compose down

# =============================================================================
# UTILITIES
# =============================================================================
.PHONY: deps-go deps-frontend deps seed

deps-go: ## Download Go dependencies
	cd $(BACKEND_DIR) && go mod download && go mod verify

deps-frontend: ## Install frontend dependencies
	cd $(FRONTEND_DIR) && npm ci

deps: deps-go deps-frontend ## Install all dependencies

seed: ## Seed database with development data
	cd $(BACKEND_DIR) && go run ./cmd/api seed

clean: ## Clean build artifacts
	rm -rf $(BINARY_DIR)
	cd $(FRONTEND_DIR) && rm -rf dist .astro

.PHONY: check
check: fmt lint typecheck test-unit ## Run all checks (format, lint, typecheck, unit tests)
	@echo "$(GREEN)All checks passed!$(RESET)"
