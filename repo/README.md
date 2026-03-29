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
- `/docs/role-matrix.md`: role visibility + endpoint policy
- `/SECURITY.md`: TLS/IP hardening and deployment notes
- `/backend/migrations/003_domains_completion.sql`: gap-closure schema

## Startup (Docker Only)
This project must be run through Docker. Local, non-Docker runs are not supported.
```bash
docker-compose up --build
```

## Service URLs
- Frontend: [http://localhost:5173](http://localhost:5173)
- Backend health: [https://localhost:8443/health](https://localhost:8443/health)
- Swagger: [https://localhost:8443/docs](https://localhost:8443/docs)
- OpenAPI YAML: [https://localhost:8443/docs/openapi.yaml](https://localhost:8443/docs/openapi.yaml)

Note: backend uses self-signed certs in `backend/certs` for local development. Trust these certs in your browser for clean HTTPS UX.

## Dev seeded accounts

For local development and testing the project includes a set of recommended dev accounts (the project seed now creates these). Use the same password for all seeded users in local/dev: **Password123!**

- **admin** / **Password123!** (super-admin)
- **admin@example.com** / **Password123!** (super-admin alias)
- **operations@example.com** / **Password123!** (operations)
- **coach@example.com** / **Password123!** (coach)
- **clinician@example.com** / **Password123!** (clinician)
- **traveler1@example.com** / **Password123!** (traveler)
- **traveler2@example.com** / **Password123!** (traveler)

These accounts are for local testing only. Do NOT use these credentials in production and remove or change them before any public deployment.

## Security Implementation
- HTTPS-only by default (`TLS_ENABLED=true`)
- Global IP allowlist via `ALLOWED_IPS`
- Optional proxy trust via `TRUST_PROXY_HEADERS`
- HSTS + X-Frame-Options + X-Content-Type-Options + Referrer-Policy
- AES-256 encryption for sensitive fields at rest
- Audit logging for role assignment changes

## Test Execution
```bash
./run_tests.sh
```
Equivalent:
```bash
docker-compose up -d --build
docker-compose exec backend ./run_tests.sh
```

Frontend tests:
```bash
cd frontend
npm run test
npm run test:e2e
```

## Additional Documentation
- Security hardening guide: [SECURITY.md](SECURITY.md)
- Role matrix: [docs/role-matrix.md](docs/role-matrix.md)
- OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml)
