# Meridian Wellness Platform Design

## 1. Purpose
Meridian Wellness is an offline-first full-stack system for wellness travel, clinic session scheduling, booking, and operational administration.

The design targets kiosk/office workflows with strong local security controls, role-based access, and predictable operations without external cloud dependencies.

## 2. High-Level Architecture

### 2.1 Components
- Frontend: React + TypeScript + Vite (`repo/frontend`)
- Backend API: Go + Echo (`repo/backend`)
- Database: PostgreSQL 16 (`docker-compose` service `db`)
- Documentation: OpenAPI + markdown docs (`docs` and `repo/backend/docs`)

### 2.2 Runtime Topology
- Browser clients call frontend on `http://localhost:5173`
- Frontend calls backend API under `/api/v1`
- Backend connects to PostgreSQL using pooled connections
- Background worker in backend processes scheduled report jobs on an interval

## 3. Backend Design

### 3.1 Layering
- API Layer: route registration, handlers, middleware (`internal/api`)
- Service Layer: auth/profile/booking workflows (`internal/service`)
- Repository Layer: SQL persistence and conflict checks (`internal/repository`)
- Platform Layer: DB connection and migrations (`internal/platform`)
- Security Layer: password policies, encryption helpers, address logic (`internal/security`)

### 3.2 Routing and Middleware
Global middleware stack:

- Panic recovery
- Request ID
- Structured request logging
- Security headers
- CORS (explicit allowed origins)
- IP allowlist (health endpoint bypass)

Versioned API group:

- Public routes: auth register/login, catalog endpoints
- Authenticated routes: profile, bookings, scheduling, community, notifications, admin, ops

Authorization model:

- JWT validates identity
- Permission middleware validates route-level access
- Ownership and staff-scope checks enforced in handlers/repository logic

## 4. Frontend Design

### 4.1 Navigation and Route Guards
Frontend routes are protected by:

- `ProtectedRoute`: requires authenticated user context
- `RoleProtectedRoute`: requires one of configured roles

Role-aware navigation matrix is defined in `src/app/roleMatrix.ts` and mirrors backend endpoint authorization.

### 4.2 UX Areas
- Core: dashboard, login, not found
- Traveler: catalog, profile, reservations
- Coach/Clinician: agenda, assigned sessions
- Operations/Admin: scheduling ops, analytics
- Admin-only: email queue, role audits, admin pages
- Shared: community and notifications

## 5. Data and Domain Model

Core domain groups:

- Identity and RBAC: users, user roles, permission audits
- Profile: addresses, contacts
- Catalog: packages, routes, hotels, attractions
- Booking/Scheduling: holds, bookings, hosts, rooms, chairs, agendas
- Community: posts, comments, follows, likes, favorites, blocks, reports
- Messaging/Notifications: in-app notifications and internal email queue
- Operations: analytics views and scheduled report jobs
- Region Service Coverage: regions, service rules, blocked postal codes

Persistence strategy:

- SQL schema managed by ordered migrations (`repo/backend/migrations`)
- Seed data loaded at startup for local/dev workflow (`repo/backend/seed/seed.sql`)

## 6. Key Functional Flows

### 6.1 Authentication and Session
1. User registers or logs in with username/password.
2. Backend enforces password policy and lockout settings.
3. JWT token is returned and attached to subsequent API calls.
4. `/auth/me` resolves current user and effective roles.

### 6.2 Address and Coverage
1. Traveler submits address.
2. Address is normalized and checked for duplicates.
3. Coverage/serviceability is computed from configured postal coverage and rules.
4. Sensitive address lines are encrypted before persistence.

### 6.3 Hold to Booking Confirmation
1. Traveler creates hold for selected package/time resources.
2. System validates multi-dimensional availability (host/room/chair/time).
3. Hold can be listed or canceled by owner.
4. Confirmation converts hold to booking using optimistic conflict protections.
5. Booking status can be advanced through controlled states.

### 6.4 Scheduling
1. Slots are generated dynamically from host, room, day, and duration.
2. Staff can read host and room agendas.
3. Ownership/scope checks prevent unauthorized cross-user reads.

### 6.5 Community and Moderation
1. Users create posts/comments and social interactions.
2. Reports are submitted through moderation endpoints.
3. Staff/admin resolve reports and actions are surfaced in user notifications.

### 6.6 Operations Reporting
1. Ops users query KPI endpoints with filters.
2. Data can be exported as CSV.
3. Scheduled report jobs are persisted.
4. Background worker processes due jobs and writes exports locally.

## 7. Security Design

### 7.1 Transport and Network
- TLS can be enforced (`TLS_ENABLED=true`)
- Startup fails if cert/key missing in TLS mode
- IP allowlist blocks non-allowed source addresses globally (except health)

### 7.2 Data Protection
- Application-level encryption for sensitive fields
- Encryption key supplied via environment
- Password hashing and lockout logic for brute-force resistance
- Structured logs with redaction and optional rotating file sink

### 7.3 API Hardening
- Explicit CORS origins only
- HSTS, frame/type/referrer hardening headers
- Centralized error responses and request tracing

## 8. Configuration Model

Important runtime configuration values include:

- `DATABASE_URL`
- `JWT_SECRET`
- `ENCRYPTION_KEY`
- `TLS_ENABLED`, `TLS_CERT_FILE`, `TLS_KEY_FILE`
- `ALLOWED_IPS`, `TRUST_PROXY_HEADERS`
- `CORS_ALLOWED_ORIGINS`
- `LOCKOUT_THRESHOLD`, `LOCKOUT_DURATION`, `TOKEN_TTL`
- `RESERVATION_HOLD`, `SLOT_GRANULARITY_MINUTES`
- `REPORT_WORKER_INTERVAL`

The backend validates critical security inputs during startup.

## 9. Deployment and Operations

### 9.1 Local Containerized Deployment
Primary development path:

`docker-compose up --build`

Services:

- `db` (Postgres)
- `backend` (Go API)
- `frontend` (React app)

### 9.2 Startup Behavior
- Backend runs migrations on startup
- Seed data is applied after migrations
- Background report processor starts after API initialization

### 9.3 Health and Docs
- Health check: `/health`
- API docs UI: `/docs`
- OpenAPI source: `/docs/openapi.yaml`

## 10. Test Strategy

- Backend API tests and unit tests under `repo/backend/tests`
- Frontend unit tests under `repo/frontend/tests`
- Frontend end-to-end tests via Playwright under `repo/frontend/e2e`
- Repository and security-specific tests near implementation packages

Recommended containerized test execution:

- Backend: `docker-compose exec backend ./run_tests.sh`
- Frontend unit: `docker-compose exec frontend npm run test`
- Frontend e2e: `docker-compose exec frontend npm run test:e2e`

## 11. Design Tradeoffs

- Offline-first operation improves resiliency in low-connectivity environments, but external integrations are intentionally limited.
- Route-level RBAC and ownership checks reduce accidental data exposure, with higher complexity in permission maintenance.
- Startup migrations simplify local consistency but require disciplined migration hygiene for production rollout.
- Background reporting in-process is simple to deploy, but very high throughput workloads may eventually require dedicated worker scaling.

## 12. Document References

- API contract: `docs/openapi.yaml`
- Security posture: `docs/SECURITY.md`
- Role access matrix: `docs/role-matrix.md`
- Functional API summary: `docs/api-spec.md`
