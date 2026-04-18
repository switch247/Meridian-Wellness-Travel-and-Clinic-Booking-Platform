# Meridian Wellness Travel & Clinic Booking Platform

## Project Overview

Offline-first full-stack platform for wellness travel and clinic session operations.

| Layer | Technology |
|---|---|
| Frontend | React 18 + MUI + React Router |
| Backend | Go 1.23 + Echo v4 |
| Database | PostgreSQL 16 |
| Auth | JWT (HS256) + bcrypt passwords |
| Runtime | Docker Compose |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser Client                        │
│    React + MUI  ·  Role-gated routes  ·  JWT in memory      │
│    /frontend/src/app    /frontend/src/pages    /src/api      │
└────────────────────────┬────────────────────────────────────┘
                         │  HTTPS / REST JSON
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Go / Echo API Server                    │
│  ┌──────────┐  ┌──────────────┐  ┌─────────────────────┐   │
│  │  Router  │  │  Middleware  │  │  Domain Handlers     │   │
│  │ /api/v1  │  │  JWT · RBAC  │  │  Auth, Booking,      │   │
│  │ + /docs  │  │  IP allow    │  │  Community, Ops,     │   │
│  │ + /health│  │  Ownership   │  │  Admin, Scheduling   │   │
│  └──────────┘  └──────────────┘  └─────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Services  ·  Repository (pgx)  ·  Security (AES-256)│   │
│  └──────────────────────────────────────────────────────┘   │
└────────────────────────┬────────────────────────────────────┘
                         │  pgx / SQL
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     PostgreSQL 16                            │
│  migrations/001–009  ·  seed.sql dev accounts               │
└─────────────────────────────────────────────────────────────┘
```

**Request flow:**  
Browser → React page calls `api/client.ts` → HTTPS to Echo router → middleware chain (IP allowlist → JWT auth → RBAC permission check → ownership guard) → handler → repository → PostgreSQL → JSON response back.

---

## Requirement Checklist

- [x] Local auth with password policy, lockout, JWT
- [x] RBAC with role-based endpoint permissions and ownership checks
- [x] Role matrix doc covering menus, APIs, and data scope
- [x] HTTPS-only backend runtime with fail-fast TLS cert/key checks
- [x] Global IP allowlist (except `/health`)
- [x] Security headers + explicit CORS origins
- [x] Address normalization, duplicate detection, coverage warnings
- [x] Catalog + booking hold flow + user hold/history endpoints
- [x] Staff APIs for users, role audits, host agenda, room agenda
- [x] Community layer (posts, threaded comments, follows/favorites/blocks/reports)
- [x] In-app notifications + internal email template queue + local CSV export
- [x] Analytics KPIs with filters, CSV export, and scheduled report jobs
- [x] Expanded catalog entities (routes/hotels/attractions) and scheduling slots API
- [x] Booking confirmation from hold with optimistic version checks
- [x] Structured logger with redaction and persistent rotating file sink
- [x] Dockerized services and root test layout

---

## Startup (Docker)

```bash
docker-compose up --build
```

## Service URLs

| Service | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend health | http://localhost:8443/health |
| Swagger UI | http://localhost:8443/docs |
| OpenAPI YAML | http://localhost:8443/docs/openapi.yaml |

> **TLS note:** Backend TLS is toggleable via `TLS_ENABLED`. Default is `false` for local compatibility. When enabled, self-signed certs in `backend/certs/` are used.

---

## Dev Seeded Accounts

All seeded users share the development password **`Password123!`**

| Username | Role |
|---|---|
| `admin` | super-admin |
| `admin@example.com` | super-admin alias |
| `operations@example.com` | operations |
| `coach@example.com` | coach |
| `clinician@example.com` | clinician |
| `traveler1@example.com` | traveler |
| `traveler2@example.com` | traveler |

> These accounts are for **local testing only**. Remove or rotate credentials before any production deployment.

---

## Test Strategy

Tests are organized in three tiers:

### Backend Unit Tests (`backend/tests/unit_tests/`)
Pure Go unit tests with no database. Cover:
- Auth service lockout logic
- Config loading and validation
- Logger structure and rotation
- Security utilities (password hashing, AES-256 encryption, address normalization)

### Backend Integration / API Tests (`backend/tests/API_tests/`)
Real HTTP integration tests that boot against a live server + database. Cover:
- Full auth flow (register, login, JWT middleware, `/auth/me`)
- Booking: hold placement, conflict detection, confirm, cancel
- Profile: addresses, contacts, ownership isolation
- Community: posts, comments, likes, follows, blocks, reports
- Notifications: list, mark-read
- Scheduling: slots, hosts, rooms, chairs, host/room agendas
- Admin: role assign, audits, regions, service rules, postal codes, catalog publish
- Ops: KPI analytics, analytics CSV export, email queue, scheduled reports
- RBAC enforcement: 401 for unauthenticated, 403 for insufficient role

### Frontend Unit Tests (`frontend/tests/*.test.{ts,tsx}`)
Vitest + React Testing Library. Cover:
- `address.ts` utility (normalization, coverage detection)
- `roleMatrix.ts` (canAccess logic for all roles)
- `AuthContext` (token lifecycle, login/logout, loading state)
- `ProtectedRoute` / `RoleProtectedRoute` (redirect behavior)
- `LoginPage` (form rendering, submit, error handling, quick register)
- `DashboardPage` (rendering, KPI cards, traveler snapshot)
- `AnalyticsPage` (schedule report action)
- `api/client.ts` (request construction, auth headers, error propagation)

### Frontend E2E Tests (`frontend/tests/e2e/*.spec.ts`)
Playwright against the Vite preview server. Cover:
- Traveler happy path: register → dashboard → reservations → catalog → community post → analytics guard

---

## Running Tests

### Local backend (no Docker required)
```bash
make test-local
```
Runs unit tests + integration tests that auto-skip when `DATABASE_URL` is unset.

### Full Docker-backed environment
```bash
./run_tests.sh
```

### Backend integration tests against live server
```bash
RUN_INTEGRATION_TESTS=true go test ./tests/... -v
```

### Frontend unit tests
```bash
cd frontend
npm test
```

### Frontend E2E tests
```bash
cd frontend
npm run test:e2e
```

Or via Docker:
```bash
docker-compose exec frontend npm run test
docker-compose exec frontend npm run test:e2e
```

---

## Security Implementation

- TLS toggleable (`TLS_ENABLED=true|false`), cert/key fail-fast on enable
- Global IP allowlist via `ALLOWED_IPS` (proxy headers gated by `TRUST_PROXY_HEADERS`)
- Explicit CORS origins + HSTS, `X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy`
- AES-256 encryption for sensitive data; rotating structured logs with redaction
- Audit trail for all role assignment changes
- See [`docs/security.md`](docs/security.md) for key generation and deployment guidance

---

## Environment Variables

```env
JWT_SECRET=your-jwt-secret
ENCRYPTION_KEY=0123456789abcdef0123456789abcdef
ALLOWED_POSTAL_CODES=10001,10002,60601,90001
CORS_ALLOWED_ORIGINS=https://localhost:5173
TLS_ENABLED=false
```

---

## Key Source Paths

| Path | Purpose |
|---|---|
| `backend/internal/api/router.go` | All route definitions and middleware wiring |
| `backend/internal/api/handlers/` | HTTP request handlers |
| `backend/internal/api/middleware/` | JWT, RBAC, IP allowlist, ownership |
| `backend/internal/repository/` | All database queries |
| `backend/internal/service/` | Business logic (auth lockout, booking holds) |
| `backend/migrations/` | SQL schema migrations (001–009) |
| `frontend/src/app/roleMatrix.ts` | Role-to-route permission map |
| `frontend/src/app/AppRoutes.tsx` | React Router definitions with role guards |
| `frontend/src/api/client.ts` | Typed API client for all backend endpoints |
| `frontend/src/context/AuthContext.tsx` | Global auth state (token, me, login, logout) |
| `docs/openapi.yaml` | Full OpenAPI 3 specification |
| `docs/role-matrix.md` | Role visibility and endpoint policy documentation |
