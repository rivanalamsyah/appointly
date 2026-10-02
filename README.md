# Appointly

> **Production-ready multi-tenant SaaS** for appointment and online booking management.  
> Initial vertical: Salon/Barbershop — architecture is vertical-agnostic.

---

## Quick Start

### Prerequisites
- Go 1.26+
- Node 20+
- Docker & Docker Compose

### Local Development

```bash
# 1. Copy environment variables
cp .env.example .env

# 2. Start infrastructure (PostgreSQL, Redis, MinIO)
make infra-up

# 3. Run database migrations
make migrate-up

# 4. Start backend API server (with live reload)
make dev-api

# 5. Start frontend dev server (in separate terminal)
make dev-frontend

# 6. Open app
# Frontend:  http://localhost:4321
# API:       http://localhost:8080
# API Docs:  http://localhost:8080/docs
# MinIO:     http://localhost:9001  (admin/minioadmin)
```

### Available Make Targets

```bash
make help              # Show all targets
make infra-up          # Start Docker infrastructure
make infra-down        # Stop Docker infrastructure
make migrate-up        # Run pending migrations
make migrate-down      # Rollback last migration
make migrate-create    # Create new migration (NAME=xxx)
make sqlc-gen          # Regenerate sqlc query code
make dev-api           # Run API server with live reload
make dev-worker        # Run background worker with live reload
make dev-frontend      # Run Astro dev server
make test-unit         # Run unit tests
make test-integration  # Run integration tests (requires infra)
make test-all          # Run all tests
make lint              # Run golangci-lint + eslint
make fmt               # Format all code
make build-api         # Build production API binary
make build-frontend    # Build production frontend
make build-all         # Build everything
make openapi-gen       # Generate OpenAPI spec
```

---

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for full architecture decisions, domain model, API design, and security architecture.

### Tech Stack Summary

| Layer | Technology |
|---|---|
| Backend | Go 1.26, Chi, sqlc, PostgreSQL 16, Redis 7 |
| Frontend | Astro 5, React 19, Tailwind CSS 4, TanStack Query |
| Background Jobs | River (Postgres-backed) |
| Auth | JWT (access) + Opaque refresh tokens |
| Storage | MinIO (dev) / S3 (prod) |
| Container | Docker + Compose |

---

## Project Structure

```
appointly/
├── backend/          # Go API service + background worker
├── frontend/         # Astro + React frontend
├── infra/            # Docker, Nginx configs
├── docker-compose.yml
├── Makefile
├── .env.example
├── README.md
└── ARCHITECTURE.md
```

---

## Environment Variables

See [.env.example](./.env.example) for all required and optional environment variables with descriptions.

---

## Domain Modules

| Module | Description |
|---|---|
| Auth | User registration, login, token refresh, logout |
| Organization | Tenant management, settings, onboarding |
| Member & RBAC | Team members, roles, permissions |
| Location | Physical locations / branches |
| Staff | Staff profiles, schedules, time-off |
| Service | Services, categories, pricing, duration |
| Resource | Bookable resources (rooms, chairs, equipment) |
| Customer | Customer profiles, history |
| Availability | Slot calculation engine |
| Appointment | Booking lifecycle management |
| Payment | Payment processing, refunds |
| Notification | Email, WhatsApp, in-app, webhook |
| Subscription | Plans, billing, usage limits |
| Audit Log | Immutable audit trail |

---

## Contributing

1. Branch from `main`: `git checkout -b feat/your-feature`
2. Follow the coding standards in [ARCHITECTURE.md](./ARCHITECTURE.md)
3. Write tests for new business logic
4. Run `make lint fmt test-unit` before committing
5. Open a PR with clear description of changes

---

## License

Proprietary — All Rights Reserved.
