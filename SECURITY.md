# Security Model & Architecture — Appointly

Appointly is designed with a **defense-in-depth, zero-trust multi-tenant security architecture**. Every component — from HTTP entry points down to database queries — enforces strict tenant isolation, authorization context verification, and audit trail immutability.

---

## 🛡️ Core Security Architecture & Threat Assumptions

### 1. Multi-Tenant Data Isolation Strategy
- **Row-Level Organization Scoping**: All domain repositories (`Appointment`, `Staff`, `Customer`, `Service`, `Location`, `Resource`, `Payment`, `Subscription`, `AuditLog`) require an explicit `organization_id` context parameter.
- **Cross-Tenant Access Rejection**: Authorization logic evaluates `authCtx.OrgID == target.OrganizationID`. Any cross-tenant attempt returns `403 Forbidden` or `404 Not Found` without disclosing the existence of another tenant's resource.
- **Systematic Verification**: Verified via automated test suite in [`cross_tenant_test.go`](file:///d:/appointly/backend/internal/usecase/security/cross_tenant_test.go).

### 2. Separation of Payment Domains
- **B2C Customer Appointment Payments**: Processed through client payment provider abstractions (e.g. Midtrans, Stripe, Cash/Pay-later).
- **B2B SaaS Subscription Payments**: Handled separately via tenant platform billing subscriptions (`Plan`, `Subscription`, `UsageRecords`).
- **Isolation Guarantee**: Client booking payments can never alter or influence tenant SaaS subscription states or vice-versa.

---

## 🔒 Authentication & Authorization (RBAC)

### 1. User Identity & Membership Separation
- **Identity Users**: System accounts holding credentials, email, password hashes, and MFA state.
- **Organization Memberships**: Connects a User to an Organization with a specific Role (`Owner`, `Admin`, `Staff`, `Member`).
- **Staff Profiles**: Operational scheduling profiles linked to an organization. A staff profile can exist without a system login account.

### 2. RBAC Permission Matrix
| Role | Manage Business & Settings | Manage Staff & Services | View All Appointments | Perform Booking & Reschedule | View Financials & Billing | View Immutable Audit Logs |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Owner** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Admin** | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Staff** | ❌ | ❌ | Assigned Only / Configurable | ✅ | ❌ | ❌ |
| **Member** | ❌ | ❌ | Self Only | ✅ | ❌ | ❌ |

---

## 📋 Immutable Audit Log System (`AuditLog`)

### 1. Immutability Guarantee
- **Append-Only Architecture**: The `AuditRepository` contract contains only `Create`, `List`, and `GetByID` methods. **No `Update` or `Delete` methods exist in the code or API endpoints.**
- **Access Restrictions**: Audit logs are readable only by authorized organization Owners and Admins (`audit:read`).

### 2. Audited System Events
- **Authentication**: `user.login`, `user.login_failed`, `user.password_changed`
- **RBAC & Memberships**: `member.role_changed`, `member.invited`, `member.removed`
- **Settings & Resources**: `org.settings_updated`, `service.created`, `staff.updated`, `location.created`
- **Transactions & Lifecycle**: `appointment.created`, `appointment.cancelled`, `payment.completed`, `refund.created`
- **SaaS Billing**: `subscription.changed`, `subscription.cancelled`

---

## 🌐 Network Security & HTTP Defense-in-Depth

### 1. HTTP Security Headers
Every response emitted by the Appointly API server includes defense-in-depth security headers configured via [`SecurityHeaders`](file:///d:/appointly/backend/internal/middleware/middleware.go):
```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; frame-ancestors 'none';
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

### 2. Rate Limiting Strategy
- **Authenticated API Routes**: 120 requests / minute per IP.
- **Public Booking Endpoints**: 60 requests / minute per IP.
- **Authentication Routes (`/api/v1/auth/*`)**: 20 requests / minute per IP (protects against brute-force attacks).

### 3. Payload Size Limits & CSRF Protection
- **Body Max Bytes**: Restricted to **2MB** per HTTP request via [`MaxBytes`](file:///d:/appointly/backend/internal/middleware/middleware.go) middleware to prevent memory exhaustion DoS.
- **CSRF Token Header**: `X-CSRF-Token` header verification for cookie-authenticated state-changing requests.

### 4. Webhook Signature Verification & Idempotency
- Inbound payment and SaaS webhooks require HMAC SHA-256 signature validation (`X-Signature` or `X-Hub-Signature-256`).
- Idempotency protection prevents duplicate event replay attacks.

---

## 🧹 Sensitive Data Protection & Error Sanitization

### 1. Password & Credentials Hashing
- Passwords are encrypted using Argon2id / Bcrypt with cryptographically random salts.
- Plaintext passwords, JWT tokens, session cookies, and payment provider secret keys are **never written to log output**.

### 2. Error Response Sanitization
- Production error responses utilize standard envelope format without exposing database schema details, SQL queries, or internal stack tracebacks.

---

## 📝 Vulnerability Reporting Policy

If you discover a security vulnerability in Appointly, please submit a report to `security@appointly.app`. We take all security vulnerabilities seriously and will respond promptly.
