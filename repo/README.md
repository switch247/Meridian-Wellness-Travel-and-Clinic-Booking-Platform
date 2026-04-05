# Meridian Wellness Travel & Clinic Booking Platform

## Project Overview
Offline-first full-stack platform for wellness travel and clinic session operations.
- Frontend: React + MUI (`/frontend`)
- Backend: Go + Echo (`/backend`)
- Database: PostgreSQL (`docker-compose`)

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

## Architecture Map
- `/backend/internal/api`: handlers, middleware, router
- `/backend/internal/repository`: DB access and query logic
- `/frontend/src/app`: routing, role guards, role matrix
- `/frontend/src/pages`: traveler/coach/ops/admin pages
- `/tests/unit_tests` and `/tests/API_tests`: executable tests
- `/docs/openapi.yaml`: API contract
- `/api/config/coverage`: backend endpoint for allowed coverage regions
  - `/docs/role-matrix.md`: role visibility + endpoint policy
- `/backend/migrations/003_domains_completion.sql`: gap-closure schema

## Startup (Docker)
Preferred full-stack runtime is Docker:
```bash
docker-compose up --build
```

## Service URLs
- Frontend: [http://localhost:5173](http://localhost:5173)
- Backend health (default): [http://localhost:8443/health](http://localhost:8443/health)
- Swagger (default): [http://localhost:8443/docs](http://localhost:8443/docs)
- OpenAPI YAML (default): [http://localhost:8443/docs/openapi.yaml](http://localhost:8443/docs/openapi.yaml)

Note: backend TLS is toggleable with `TLS_ENABLED`. Default is disabled for local test compatibility. If enabled, backend uses self-signed certs in `backend/certs` for local development.

## Dev seeded accounts

For local development and testing the project includes a set of recommended dev accounts (the project seed now creates these). All seeded users use the same development password: **Password123!**

Common users (quick reference):

- **admin** — Password: **Password123!** (super-admin)
- **admin@example.com** — Password: **Password123!** (super-admin alias)
- **operations@example.com** — Password: **Password123!** (operations)
- **coach@example.com** — Password: **Password123!** (coach)
- **clinician@example.com** — Password: **Password123!** (clinician)
- **traveler1@example.com** — Password: **Password123!** (traveler)
- **traveler2@example.com** — Password: **Password123!** (traveler)

These accounts are for local testing only. Do NOT use these credentials in production and remove or change them before any public deployment.

## Security Implementation
- TLS is supported and toggleable (`TLS_ENABLED=true|false`), with cert/key fail-fast enforcement when enabled
- Global IP allowlist via `ALLOWED_IPS` (proxy headers gated by `TRUST_PROXY_HEADERS`)
- Explicit CORS origins plus HSTS, `X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy`
- AES-256 encryption for sensitive data and rotating logs with redaction
- Audit logging for role assignment changes
- See [`docs/security.md`](docs/security.md) for deployment and key-generation steps

### Sample Postal Coverage & Env Example

For local testing of address coverage and normalization, you can provide a sample CSV in the `ALLOWED_POSTAL_CODES` environment (or use `ALLOWED_IPS` style variables). Example `.env` snippet:

```
JWT_SECRET=your-jwt-secret
ENCRYPTION_KEY=0123456789abcdef0123456789abcdef
ALLOWED_POSTAL_CODES=10001,10002,60601,90001
CORS_ALLOWED_ORIGINS=https://localhost:5173
TLS_ENABLED=false
```

The backend `config` loads `AllowedPostalCode` from a compile-time default; supplying `ALLOWED_POSTAL_CODES` lets you test coverage warnings for out-of-service postal codes.

## Test Execution
Local-first backend verification (no Docker required):
```bash
make test-local
```

This runs `backend` tests directly and is the primary audit path in constrained environments. DB integration tests auto-skip when `DATABASE_URL` is unset.

Docker-backed full environment tests:
```bash
./run_tests.sh
```

Frontend tests (manual):
```bash
docker-compose exec frontend npm run test
docker-compose exec frontend npm run test:e2e
```
