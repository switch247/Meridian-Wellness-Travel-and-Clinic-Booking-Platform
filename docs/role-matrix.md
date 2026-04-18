# Role Matrix

This document summarizes role visibility and endpoint permission intent.

## Roles
- traveler
- coach
- clinician
- operations
- admin

## UI Visibility Summary
- traveler: catalog, profile, reservations, community, notifications
- coach/clinician: agenda, assigned sessions, community, notifications
- operations: scheduling, analytics, reports, email queue, community
- admin: all operations capabilities plus role and policy administration

## API Permission Source of Truth
Permission checks are enforced in backend middleware and route registration.

- Permission constants and role mapping:
  - backend/internal/api/middleware/permissions.go
- Route-to-permission wiring:
  - backend/internal/api/router.go

## Notes
- Object-level and tenant/location checks are additionally enforced in domain handlers and repository methods.
- This matrix is intentionally concise; implementation files are authoritative.
