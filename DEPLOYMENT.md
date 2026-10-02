# 🚀 Appointly — Production Deployment & Operational Runbook

This document serves as the authoritative operational guide for deploying, managing, and maintaining the **Appointly Multi-Tenant Appointment Booking SaaS** in Staging and Production environments.

---

## 1. System Architecture & Topology

```
                         ┌────────────────────────────────┐
                         │   TLS / HTTPS (Port 80/443)    │
                         │   Caddy / Nginx Reverse Proxy   │
                         └───────────────┬────────────────┘
                                         │
                 ┌───────────────────────┴───────────────────────┐
                 │                                               │
                 ▼                                               ▼
   ┌───────────────────────────┐                   ┌───────────────────────────┐
   │    Astro Frontend App     │                   │       Go API Server       │
   │  Node Standalone Runner   │                   │     Chi / Slog Runtime    │
   │        (Port 4321)        │                   │        (Port 8080)        │
   └───────────────────────────┘                   └─────────────┬─────────────┘
                                                                 │
                                     ┌───────────────────────────┼───────────────────────────┐
                                     │                           │                           │
                                     ▼                           ▼                           ▼
                       ┌──────────────────────────┐  ┌───────────────────────┐  ┌──────────────────────────┐
                       │  PostgreSQL 16 (Primary) │  │  Redis 7 (Cache/RL)   │  │  Object Storage (S3/GCS) │
                       │    Multi-Tenant Data     │  │   Sessions & Limits   │  │   Customer/Staff Media   │
                       └──────────────────────────┘  └───────────────────────┘  └──────────────────────────┘
```

---

## 2. Environment Variables & Secret Management

### Critical Production Rules
1. **Never commit `.env` or plain-text secrets to version control**.
2. Inject secrets dynamically at runtime via Cloud Provider Secret Managers (AWS Secrets Manager, GCP Secret Manager, Vault, or GitHub Repository Secrets).
3. Minimum secret requirements:
   - `JWT_SECRET`: Random 32+ byte string (e.g. `openssl rand -base64 32`).
   - `DB_PASSWORD`: Strong password, minimum 20 characters.
   - `REDIS_PASSWORD`: Strong password, minimum 20 characters.

### Environment Parameter Matrix

| Variable | Recommended Production Value | Description |
| :--- | :--- | :--- |
| `APP_ENV` | `production` | Enables production security & disables debug endpoints. |
| `PORT` | `8080` | Internal HTTP listening port for Go API server. |
| `DB_HOST` | `postgres` (or AWS RDS endpoint) | PostgreSQL database hostname. |
| `DB_PORT` | `5432` | PostgreSQL database port. |
| `DB_NAME` | `appointly_prod` | Production database name. |
| `DB_USER` | `appointly` | Master DB user with tenant access privileges. |
| `DB_PASSWORD` | `<SECRET>` | Production database password. |
| `DB_SSL_MODE` | `require` / `verify-full` | Enforces TLS encryption for database connections. |
| `REDIS_HOST` | `redis` (or ElastiCache endpoint) | Redis server hostname. |
| `REDIS_PORT` | `6379` | Redis server port. |
| `REDIS_PASSWORD` | `<SECRET>` | Redis authentication password. |
| `JWT_SECRET` | `<SECRET>` | Minimum 32-character string for HMAC JWT signatures. |
| `CORS_ALLOWED_ORIGINS` | `https://appointly.com,https://*.appointly.com` | Allowed CORS origins for web applications. |
| `LOG_FORMAT` | `json` | Enforces structured JSON logging for Datadog / CloudWatch. |
| `LOG_LEVEL` | `info` | Minimum log verbosity level (`info`, `warn`, `error`). |

---

## 3. Containerization & Orchestration

### Multi-Stage Docker Builds
Both frontend and backend utilize multi-stage Docker builds to ensure minimal production image footprints, security isolation, and fast startup:
- **Backend Dockerfile**: `golang:1.24-alpine` builder -> `alpine:3.21` runner with non-root user `appointly:10001` (Binary size ~18MB).
- **Frontend Dockerfile**: `node:22-alpine` builder -> `node:22-alpine` runner with non-root user `appointly:10001`.

### Production Deployment Commands

#### Staging Environment Setup
```bash
# 1. Clone repository
git clone https://github.com/rivanalamsyah/appointly.git
cd appointly

# 2. Prepare environment variables
cp .env.example .env
# Edit .env with staging credentials

# 3. Build & start containers
docker compose -f docker-compose.prod.yml up -d --build

# 4. Verify deployment health
docker compose -f docker-compose.prod.yml ps
curl -i http://localhost/healthz
```

#### Production Rolling Deployment
```bash
# Pull latest code
git pull origin main

# Build updated Docker images without downtime
docker compose -f docker-compose.prod.yml build --no-cache

# Run database migrations first
docker compose -f docker-compose.prod.yml run --rm migrations

# Perform zero-downtime rolling restart
docker compose -f docker-compose.prod.yml up -d --no-deps --scale backend=2 --no-recreate
docker compose -f docker-compose.prod.yml up -d --no-deps backend frontend proxy
```

---

## 4. Database Migration Strategy

Database migrations are stored in `backend/db/migrations/` using timestamp/sequence versioning (`000001` to `000008`).

### Migration Execution Guidelines
1. **Backward Compatibility**: Every migration MUST be backward-compatible with the currently running application code.
2. **Column Additions**: Add new columns as `NULL` or with explicit default values to prevent breaking active queries.
3. **Execution Tool**: Use `migrate/migrate:v4.17.0` container inside CI/CD or production orchestration before starting new application pods.

#### Manual Migration Commands
```bash
# Run all pending up migrations
docker run --rm -v $(pwd)/backend/db/migrations:/migrations migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
  up

# Check current migration version
docker run --rm -v $(pwd)/backend/db/migrations:/migrations migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
  version
```

---

## 5. Health, Readiness, and Graceful Shutdown

### Health & Readiness Probes
The Go backend exposes standard health endpoints:
- `GET /health` / `GET /healthz` / `GET /livez`: Returns HTTP `200 OK` with JSON `{"status":"ok","version":"x.y.z"}` if application process is running. Used for Liveness probes.
- `GET /ready` / `GET /readyz`: Returns HTTP `200 OK` with JSON `{"status":"ready"}` if database and Redis connectivity are functional. Used for Readiness probes.

### Graceful Shutdown Specification
The Go server listens for `SIGINT` and `SIGTERM` signals:
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
```
Upon receiving a signal:
1. The server stops accepting new incoming connections immediately.
2. Active HTTP requests are given a grace period of `30s` (`API_SHUTDOWN_TIMEOUT`) to finish execution cleanly.
3. Database connections and Redis pools are closed after active HTTP requests complete.

---

## 6. Worker & Background Job Idempotency

1. **At-Least-Once Delivery**: Background workers (webhook notifications, email reminders, audit logging) operate under at-least-once delivery guarantees.
2. **Idempotency Safeguards**:
   - Webhook processing checks `webhook_deliveries.id` or external provider transaction IDs before applying side effects.
   - Payment webhooks utilize idempotent upsert transactions on `payments` and `appointments`.

---

## 7. Observability & Logging Standard

1. **Structured JSON Logging**: All log outputs use `log/slog` formatted as JSON strings for automated ingestion into Datadog, AWS CloudWatch, or Grafana Loki.
2. **Traceability**: Every HTTP request is annotated with a unique `X-Request-ID` header propagating across middleware, application logs, and database queries.
3. **Log Fields**:
   ```json
   {
     "time": "2026-10-03T02:40:00Z",
     "level": "INFO",
     "msg": "appointment created",
     "request_id": "req-987123",
     "org_id": "c001-org",
     "appointment_id": "appt-456"
   }
   ```

---

## 8. Database Backup & Disaster Recovery Strategy

### Automated Backup Schedule
- **Daily Full Backups**: Conducted every night at 02:00 UTC using `pg_dump` with custom compressed format.
- **Transaction Logs (WAL)**: Archive Continuous Archiving (WAL-G / pgBackRest) to AWS S3 bucket with 30-day retention policy for Point-In-Time Recovery (PITR).

#### Manual Backup Command
```bash
docker exec -t appointly-prod-postgres pg_dump -U appointly -d appointly_prod -Fc > backup_$(date +%Y%m%d_%H%M%S).dump
```

#### Restore Procedure
```bash
# 1. Create target database
docker exec -it appointly-prod-postgres createdb -U appointly appointly_restore

# 2. Restore schema and data from dump
docker exec -i appointly-prod-postgres pg_restore -U appointly -d appointly_restore < backup_20261003_020000.dump
```

---

## 9. Zero-Downtime Rollback Strategy

If a deployment failure occurs:

1. **Automated Rollback via Docker Compose**:
   ```bash
   # Roll back container images to previous tagged release
   docker compose -f docker-compose.prod.yml down
   git checkout tags/v1.0.0
   docker compose -f docker-compose.prod.yml up -d
   ```
2. **Database Rollback**:
   If a migration needs to be rolled back:
   ```bash
   docker run --rm -v $(pwd)/backend/db/migrations:/migrations migrate/migrate:v4.17.0 \
     -path=/migrations \
     -database="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
     down 1
   ```

---

## 10. Final Pre-Production Readiness Review Checklist

- [x] **Security Headers**: `SecurityHeaders` middleware active (CSP, HSTS, X-Content-Type-Options).
- [x] **Rate Limiting**: Per-route rate limits enforced (`RateLimit` & `PublicRateLimit`).
- [x] **Database Indexes**: FK indexes (`payment_id`, `org_id`, `staff_id`, `customer_id`) and double-booking unique index (`idx_appointments_no_double_book`) validated.
- [x] **CORS Configuration**: Restricted to explicit domain whitelist (`CORS_ALLOWED_ORIGINS`).
- [x] **Non-Root Containers**: Both frontend and backend execution users restricted to non-root system ID `10001`.
- [x] **CI/CD Pipeline**: GitHub Actions `.github/workflows/ci.yml` configured to enforce formatting, linting, typechecking, unit tests, integration tests, and image build verification.
