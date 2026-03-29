# Role Matrix

## Roles
- traveler
- coach
- clinician
- operations
- admin

Coach and clinician share the same permission group in this release.

## Menu Visibility
| Menu/Page | traveler | coach/clinician | operations | admin |
|---|---|---|---|---|
| Dashboard | yes | yes | yes | yes |
| Catalog | yes | no | yes | yes |
| Community | yes | yes | yes | yes |
| Notifications | yes | yes | yes | yes |
| Booking | yes | no | no | yes |
| Profile | yes | no | no | yes |
| My Reservations | yes | no | no | no |
| My Agenda | no | yes | no | no |
| Assigned Sessions | no | yes | no | no |
| Scheduling Ops | no | no | yes | yes |
| Analytics | no | no | yes | yes |
| Email Queue | no | no | yes | yes |
| Role Audits | no | no | yes | yes |
| Admin | no | no | yes | yes |
| API Docs | no | no | yes | yes |

## Endpoint Permissions
| Endpoint | traveler | coach/clinician | operations | admin | Data scope |
|---|---|---|---|---|---|
| GET /api/v1/auth/me | yes | yes | yes | yes | self |
| GET /api/v1/catalog (+ routes/hotels/attractions) | yes | no | yes | yes | published only |
| GET/POST/DELETE /api/v1/profile/addresses | yes | no | no | traveler self | self |
| POST/GET/DELETE /api/v1/bookings/holds | yes | no | no | traveler self | self |
| POST /api/v1/bookings/confirm | yes | no | no | yes | hold owner only |
| GET /api/v1/bookings/history | yes | no | no | traveler self | self |
| GET /api/v1/scheduling/slots | yes | no | yes | yes | filtered by query |
| GET /api/v1/scheduling/hosts/:id/agenda | own id | own id | any | any | ownership for non-staff |
| GET /api/v1/scheduling/rooms/:id/agenda | no | no | any | any | staff scope |
| Community read/write endpoints | yes | yes | yes | yes | subject to user block/report rules |
| GET/POST /api/v1/notifications | yes | yes | yes | yes | self |
| GET /api/v1/admin/users | no | no | yes | yes | staff scope |
| POST /api/v1/admin/roles/assign | no | no | yes | yes | audited |
| GET /api/v1/admin/roles/audits | no | no | yes | yes | staff scope |
| POST /api/v1/admin/reports/:id/resolve | no | no | yes | yes | staff scope |
| GET /api/v1/ops/analytics/kpis | no | no | yes | yes | staff scope |
| GET /api/v1/ops/analytics/export | no | no | yes | yes | staff scope |
| POST /api/v1/ops/reports/schedule | no | no | yes | yes | staff scope |
| GET/POST /api/v1/ops/email/queue | no | no | yes | yes | staff scope |
| POST /api/v1/ops/email/export | no | no | yes | yes | staff scope |

## Ownership Rules
- Travelers can only read/change their own addresses, holds, booking confirmations, and notifications.
- Coach/clinician can read own agenda only unless elevated staff role is also assigned.
- Operations/admin can access cross-user operational views.
