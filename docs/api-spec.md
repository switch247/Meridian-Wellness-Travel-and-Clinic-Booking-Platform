# Meridian Wellness API Specification

## Overview
Meridian Wellness is an offline-first wellness travel and clinic booking platform.

- Protocol: HTTPS in production-like setups
- API style: REST over JSON
- Version prefix: `/api/v1`
- Auth: JWT bearer token
- Primary actors: traveler, coach, clinician, operations, admin

This document is a practical API guide derived from the current OpenAPI contract and backend router implementation.

## Base URLs

- Local backend root: `https://localhost:8443`
- API base: `https://localhost:8443/api/v1`
- Swagger UI: `https://localhost:8443/docs`
- OpenAPI YAML: `https://localhost:8443/docs/openapi.yaml`

## Authentication

### Register
- Method: `POST`
- Path: `/api/v1/auth/register`
- Access: Public
- Purpose: Create a local account

### Login
- Method: `POST`
- Path: `/api/v1/auth/login`
- Access: Public
- Purpose: Authenticate and issue JWT

### Current User
- Method: `GET`
- Path: `/api/v1/auth/me`
- Access: Authenticated
- Purpose: Return current user profile and role context

### Authorization Header
For authenticated endpoints, send:

`Authorization: Bearer <token>`

## Security and Access Controls

- JWT-protected route group under `/api/v1`
- Permission checks enforced per endpoint in middleware
- Ownership constraints enforced for self-scoped resources (addresses, contacts, holds, notifications)
- IP allowlist enforcement at middleware layer (except `/health`)
- CORS restricted to configured origins
- Security headers are added globally

## Domain Endpoints

### Catalog

- `GET /api/v1/catalog` - Published packages and pricing calendar
- `GET /api/v1/config/coverage` - Allowed coverage regions list
- `GET /api/v1/catalog/routes` - Published travel routes
- `GET /api/v1/catalog/hotels` - Published hotels
- `GET /api/v1/catalog/attractions` - Published attractions

### Profile

- `GET /api/v1/profile/addresses` - List traveler addresses
- `POST /api/v1/profile/addresses` - Add traveler address
- `DELETE /api/v1/profile/addresses/{id}` - Delete address
- `GET /api/v1/profile/contacts` - List emergency/profile contacts
- `POST /api/v1/profile/contacts` - Add contact
- `DELETE /api/v1/profile/contacts/{id}` - Delete contact

Address handling includes normalization, duplicate detection, coverage checks, and encrypted storage for sensitive lines.

### Booking

- `POST /api/v1/bookings/holds` - Create reservation hold
- `GET /api/v1/bookings/holds` - List active holds
- `DELETE /api/v1/bookings/holds/{id}` - Cancel hold
- `POST /api/v1/bookings/confirm` - Confirm hold into booking
- `GET /api/v1/bookings/history` - Booking history
- `POST /api/v1/bookings/{id}/status` - Update booking status and optional notes

Expected conflict responses:

- `409 Conflict` for optimistic-lock or state conflicts during hold/confirm flows

### Scheduling

- `GET /api/v1/scheduling/slots` - Dynamic slot generation
	- Required query params: `hostId`, `roomId`, `day`, `duration`
	- Allowed duration values: `30`, `45`, `60`
- `GET /api/v1/scheduling/hosts` - Host list for scheduling context
- `GET /api/v1/scheduling/rooms` - Room list
- `GET /api/v1/scheduling/rooms/{id}/chairs` - Chairs in room
- `GET /api/v1/scheduling/hosts/{id}/agenda` - Host agenda
- `GET /api/v1/scheduling/rooms/{id}/agenda` - Room agenda

### Community

- `GET /api/v1/community/posts`
- `POST /api/v1/community/posts`
- `GET /api/v1/community/posts/{id}/comments`
- `POST /api/v1/community/posts/{id}/comments`
- `POST /api/v1/community/favorites`
- `POST /api/v1/community/likes`
- `POST /api/v1/community/follows`
- `POST /api/v1/community/blocks`
- `POST /api/v1/community/reports`

### Notifications

- `GET /api/v1/notifications` - Notification feed for current user
- `POST /api/v1/notifications/{id}/read` - Mark one notification as read

### Admin

- `POST /api/v1/admin/roles/assign` - Assign role (audited)
- `GET /api/v1/admin/users` - List users with role filtering context
- `GET /api/v1/admin/roles/audits` - Permission audit feed
- `POST /api/v1/admin/reports/{id}/resolve` - Resolve moderation report
- `GET /api/v1/admin/regions` - Region catalog
- `POST /api/v1/admin/regions` - Create region
- `POST /api/v1/admin/regions/{id}/service-rule` - Upsert service rule
- `GET /api/v1/admin/service/blocked-postal-codes` - List blocked postal codes
- `POST /api/v1/admin/service/blocked-postal-codes` - Add blocked postal code

### Operations

- `GET /api/v1/ops/analytics/kpis` - KPI metrics
- `GET /api/v1/ops/analytics/export` - Export analytics CSV
- `GET /api/v1/ops/reports` - Scheduled report jobs
- `POST /api/v1/ops/reports/schedule` - Schedule local report generation
- `GET /api/v1/ops/email/queue` - List internal email queue
- `POST /api/v1/ops/email/queue` - Queue an email template payload
- `POST /api/v1/ops/email/export` - Export queue as CSV

## Common Status Codes

- `200 OK` - Successful read/update action
- `201 Created` - Resource created
- `400 Bad Request` - Invalid payload or query
- `401 Unauthorized` - Missing/invalid token
- `403 Forbidden` - Permission or ownership denied
- `404 Not Found` - Resource does not exist
- `409 Conflict` - Version/state conflict

## Role Scope Summary

- traveler: self-service flows (profile, holds, booking history, notifications, community)
- coach/clinician: agenda-centric access and shared community/notification access
- operations: cross-user operations views, analytics, scheduling, moderation/admin support
- admin: full platform governance, role assignment, region/service-rule management

Detailed role/page/endpoint mapping is maintained in `repo/docs/role-matrix.md`.

## Example Requests

### Login
```http
POST /api/v1/auth/login HTTP/1.1
Host: localhost:8443
Content-Type: application/json

{
	"username": "traveler1@example.com",
	"password": "Password123!"
}
```

### Create Hold
```http
POST /api/v1/bookings/holds HTTP/1.1
Host: localhost:8443
Authorization: Bearer <token>
Content-Type: application/json

{
	"packageId": 1,
	"hostId": 2,
	"roomId": 1,
	"chairId": 3,
	"startAt": "2026-04-03T10:00:00Z",
	"duration": 60
}
```

### Fetch Dynamic Slots
```http
GET /api/v1/scheduling/slots?hostId=2&roomId=1&day=2026-04-03&duration=60 HTTP/1.1
Host: localhost:8443
Authorization: Bearer <token>
```

## Implementation Notes

- The canonical machine-readable contract is `repo/backend/docs/openapi.yaml` (also served at runtime under `/docs/openapi.yaml`).
- Some endpoint payload details are intentionally concise in YAML and are expanded in handler/repository logic.
- Booking status supports: `scheduled`, `confirmed`, `checked_in`, `in_progress`, `completed`, `cancelled`.
