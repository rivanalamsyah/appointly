# Appointly — Architecture Decision Record

> **Version**: 0.1.0  
> **Status**: Living Document  
> **Last Updated**: 2026-10-02  

---

## 1. System Overview

**Appointly** is a production-ready, multi-tenant SaaS platform for appointment and online booking management. The initial vertical is salon/barbershop, but the architecture is **vertical-agnostic** — the same core can support clinics, consultants, lawyers, tutors, trainers, and other service businesses without redesigning the core database or booking engine.

### Core Principles

| Principle | Implementation |
|---|---|
| Multi-tenancy | Organization = tenant; every business record is tenant-isolated |
| Vertical-agnostic | Generic `staff`, `service`, `resource` model; no vertical-specific tables in core |
| Domain-driven | Clear domain boundaries; no shared mutable state between domains |
| Layered architecture | Handler → Use Case → Domain → Repository |
| Security-first | Auth, RBAC, CSRF, rate-limiting, audit log built-in from day one |
| Observability | Structured logging, request IDs, health/readiness endpoints from day one |

---

## 2. Technology Stack

### Backend
| Concern | Technology | Rationale |
|---|---|---|
| Language | Go 1.26+ | Performance, simplicity, strong stdlib, excellent concurrency |
| HTTP Router | Chi | Lightweight, idiomatic, middleware-composable |
| Database | PostgreSQL 16 | ACID, rich constraint system, JSONB where needed |
| ORM/Query | sqlc + raw SQL | Type-safe queries without ORM magic, full control |
| Migrations | golang-migrate | Reversible numbered migrations, CLI + embedded |
| Cache/Queue | Redis 7 | Session store, distributed lock, pub/sub, background jobs |
| Object Storage | MinIO (dev) / S3 (prod) | Blob storage for uploads/avatars |
| Auth Tokens | JWT (access) + opaque refresh tokens | Stateless access, revocable refresh |
| Password Hash | Argon2id | Memory-hard, OWASP recommended |
| Background Jobs | River (Go) | Postgres-backed job queue, durable, transactional enqueue |
| API Spec | OpenAPI 3.1 | Contract-first documentation |
| Container | Docker + Compose | Consistent dev/prod parity |

### Frontend
| Concern | Technology | Rationale |
|---|---|---|
| Framework | Astro 5 | Islands architecture, partial hydration, fast static output |
| Interactive | React 19 | Only where complex interaction needed (calendar, wizard, modals) |
| Styling | Tailwind CSS 4 | Utility-first, consistent design tokens |
| Server State | TanStack Query v5 | Cache, background refetch, optimistic updates |
| Forms | React Hook Form + Zod | Schema-driven validation, minimal re-renders |
| Type Safety | TypeScript 5.5 | End-to-end type safety |
| HTTP Client | ky (fetch wrapper) | Tiny, interceptor-aware |

---

## 3. Repository Structure

```
appointly/
├── backend/                    # Go service
│   ├── cmd/
│   │   ├── api/               # API server entrypoint
│   │   └── worker/            # Background worker entrypoint
│   ├── internal/
│   │   ├── config/            # Configuration loading (env-based)
│   │   ├── server/            # HTTP server setup, middleware stack
│   │   ├── domain/            # Domain models & interfaces (pure Go)
│   │   │   ├── auth/
│   │   │   ├── organization/
│   │   │   ├── member/
│   │   │   ├── rbac/
│   │   │   ├── location/
│   │   │   ├── staff/
│   │   │   ├── service/
│   │   │   ├── resource/
│   │   │   ├── customer/
│   │   │   ├── scheduling/
│   │   │   ├── appointment/
│   │   │   ├── payment/
│   │   │   ├── notification/
│   │   │   ├── subscription/
│   │   │   ├── audit/
│   │   │   └── webhook/
│   │   ├── usecase/           # Application use cases (orchestration)
│   │   ├── handler/           # HTTP handlers (transport layer)
│   │   │   └── v1/
│   │   │       └── public/    # Public booking API
│   │   ├── repository/        # Database implementations
│   │   │   ├── postgres/
│   │   │   └── redis/
│   │   ├── worker/            # Background job handlers
│   │   ├── middleware/        # HTTP middleware
│   │   └── pkg/               # Internal shared packages
│   │       ├── apperror/
│   │       ├── logger/
│   │       ├── validator/
│   │       ├── pagination/
│   │       ├── idgen/
│   │       ├── timeutil/
│   │       └── crypto/
│   ├── db/
│   │   ├── migrations/        # golang-migrate SQL files
│   │   └── queries/           # sqlc .sql query files
│   ├── gen/                   # sqlc generated code (committed)
│   ├── api/
│   │   └── openapi.yaml       # OpenAPI 3.1 spec
│   ├── sqlc.yaml
│   ├── go.mod
│   └── go.sum
│
├── frontend/                  # Astro application
│   ├── src/
│   │   ├── pages/
│   │   │   ├── index.astro           # Marketing home
│   │   │   ├── features.astro
│   │   │   ├── pricing.astro
│   │   │   ├── login.astro
│   │   │   ├── register.astro
│   │   │   ├── dashboard/
│   │   │   └── book/[slug]/
│   │   ├── components/
│   │   │   ├── marketing/
│   │   │   ├── layout/
│   │   │   ├── ui/
│   │   │   └── features/
│   │   ├── lib/
│   │   │   ├── api/
│   │   │   ├── auth/
│   │   │   └── utils/
│   │   ├── types/
│   │   └── styles/
│   ├── astro.config.mjs
│   ├── tailwind.config.mjs
│   ├── tsconfig.json
│   └── package.json
│
├── infra/
│   ├── docker/
│   └── nginx/
│
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
├── .gitignore
├── Makefile
├── README.md
└── ARCHITECTURE.md
```

---

## 4. Domain Model

### 4.1 Core Entities

```
Organization (tenant)
  ├── has many Location
  ├── has many OrganizationMember → User
  ├── has many Staff
  ├── has many ServiceCategory
  ├── has many Service
  ├── has many Resource
  ├── has many Customer
  ├── has many Appointment
  ├── has many Webhook
  └── has one Subscription → Plan

User (identity/account)
  └── has many OrganizationMember

Staff (business entity, linked to User optionally)
  ├── belongs to Organization
  ├── optionally linked to User via user_id
  ├── works at many Location (staff_locations)
  └── provides many Service (staff_services)

Service
  ├── belongs to Organization
  ├── belongs to ServiceCategory
  ├── has duration, price, buffer_before, buffer_after
  └── assigned to many Staff (staff_services)

Resource (room, chair, equipment)
  ├── belongs to Organization
  ├── belongs to Location
  └── can be required by Service (service_resources)

Appointment
  ├── belongs to Organization
  ├── belongs to Location
  ├── belongs to Service
  ├── belongs to Staff
  ├── optionally belongs to Customer
  ├── optionally belongs to Resource
  ├── has status: pending|confirmed|rescheduled|completed|cancelled|no_show
  └── has many AppointmentStatusHistory

Payment
  ├── belongs to Appointment
  ├── has status: pending|paid|failed|refunded|partially_refunded
  └── has many Refund
```

### 4.2 Tenant Isolation Rules

1. Every business entity carries `organization_id` (NOT NULL, FK).
2. All repository queries MUST include `organization_id` from authenticated context.
3. `organization_id` from client request body is **IGNORED** — it comes from the JWT/session.
4. Row-level checks in critical paths use DB-level constraints.

### 4.3 Availability Engine

```
Input:
  - organization_id, service_id, staff_id (optional), location_id
  - date_range, timezone

Algorithm:
  1. Load BusinessHours for location on requested date
  2. Load StaffSchedule for each candidate staff on date
  3. Load StaffTimeOff overlapping date range
  4. Load existing Appointments (status: pending|confirmed|rescheduled)
     with service duration + buffer_before + buffer_after
  5. Load Resource availability if service requires resource
  6. Apply BookingSettings (min_advance_notice, max_advance_days, slot_granularity)
  7. Generate candidate slots at configured granularity
  8. For each candidate slot: check all constraints → available/unavailable

Race condition protection:
  - SELECT ... FOR UPDATE on (staff_id, start_time) before inserting
  - Unique partial index on appointments prevents double-booking at DB level
```

---

## 5. API Design

### URL Structure
```
/api/v1/                          # Authenticated API
/api/v1/public/                   # Public (no auth, rate-limited)
/health                           # Health check
/ready                            # Readiness probe
```

### Auth Strategy
- **Access Token**: JWT, 15 min TTL, HS256/RS256
- **Refresh Token**: Opaque, stored in DB, httpOnly cookie
- **CSRF**: Double-submit cookie for cookie-based auth
- **Tenant context**: Extracted from JWT claims, never from request body

### Standard Response Envelope
```json
{ "data": {...}, "meta": {"request_id": "...", "timestamp": "..."} }
{ "data": [...], "meta": {"pagination": {"page":1,"per_page":20,"total":150}} }
{ "error": {"code": "VALIDATION_FAILED", "message": "...", "details": [...]} }
```

---

## 6. Security Architecture

| Layer | Control |
|---|---|
| Transport | TLS enforced in production |
| Authentication | JWT + Refresh Token rotation |
| Authorization | RBAC + Permission checks in use case layer |
| Tenant Isolation | organization_id from context, never from client |
| Rate Limiting | Token bucket per IP + per user |
| CORS | Strict allowlist |
| Input Validation | Schema validation, parameterized queries only |
| Password | Argon2id |
| Audit Log | Immutable append-only log |
| Webhooks | HMAC-SHA256 signature verification |

---

## 7. Background Job Architecture

Jobs are enqueued transactionally using **River** (Postgres-backed):

| Job | Trigger | Retry |
|---|---|---|
| `notification.email` | appointment/payment events | 5x exp backoff |
| `notification.whatsapp` | appointment events | 3x exp backoff |
| `notification.webhook` | all published events | 10x exp backoff |
| `appointment.reminder` | cron, 24h/1h before | 3x |
| `report.generate` | on-demand / scheduled | 2x |

---

## 8. Decision Log

| # | Decision | Rationale | Alternatives |
|---|---|---|---|
| 1 | Chi over Gin/Echo | Stdlib-compatible, minimal, composable | Gin, Echo |
| 2 | sqlc over GORM | Type-safe, no reflection, explicit SQL | GORM, squirrel |
| 3 | River for jobs | Transactional enqueue on Postgres | Asynq (Redis-only) |
| 4 | Argon2id for passwords | OWASP recommended, memory-hard | bcrypt |
| 5 | Single-schema multi-tenancy | Simpler ops, easier migrations | Per-tenant schema |
| 6 | UUIDv7 for public IDs | Time-ordered, K-sortable | Auto-increment, UUIDv4 |
| 7 | Astro Islands | Partial hydration, best perf | Next.js, pure SPA |
| 8 | Selective soft delete | Only where audit/recovery needed | Universal soft delete |
| 9 | Availability Engine | Stateless candidate slot pipeline; atomic lock at booking | Hardcoded static 30m slots |

---

## 9. Availability Engine Architecture & Pipeline Algorithm

The **Availability Engine** is a high-performance, stateless application service (`internal/usecase/availability`) designed to compute valid candidate booking slots dynamically.

### Key Design Guarantee
- **Dynamic Slot Generation**: Slot starts and durations are computed on-the-fly based on `service.DurationMinutes` + `service.BufferBefore` + `service.BufferAfter` stepped by configurable `organization.Settings.SlotGranularityMinutes` (e.g., 15m, 30m, 60m). Static hard-coded slot grids are strictly avoided.
- **Candidate Non-Reservation**: Availability slots are stateless candidate suggestions. To prevent race conditions, the final appointment creation step performs an atomic transaction lock (`SELECT FOR UPDATE`) on the target staff and resource window.

### Pipeline Algorithm Phases

```
[Query Input] (org, location, service, staff, date_range, timezone)
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Service Status & Duration Calculation              │
│ - Verify service.Status == ACTIVE                          │
│ - totalSlotDuration = BufferBefore + Duration + BufferAfter │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Organization Settings & Advance Notice Bounds      │
│ - earliestAllowed = now + MinAdvanceBookingHours            │
│ - latestAllowed   = now + MaxAdvanceDays                    │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Staff Candidates & Location Intersections          │
│ - Filter ACTIVE staff assigned to the requested Service     │
│ - Intersect Location Business Hours with Staff Shift Hours │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 4: Time Window Subtractions (Pure Set Math)            │
│ - Subtract Staff Breaks from shift windows                  │
│ - Subtract Approved Staff Time-Off entries                  │
│ - Subtract Existing Appointments + Buffer Windows           │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 5: Resource Constraints & Granular Slot Step          │
│ - Verify at least 1 assigned Resource is ACTIVE & free      │
│ - Iterate t in steps of SlotGranularityMinutes              │
│ - Filter slots within [earliestAllowed, latestAllowed]      │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
[Formatted & Chronologically Sorted Slots Output]
```

