# Role Matrix & Scope

## Roles
- **Traveler**: books travel/wellness experiences and clinician slots, manages contacts/addresses, views bookings/holds, interacts with community.
- **Coach/Clinician**: views assigned sessions, checks in/out, logs clinical notes (encrypted), participates in community moderation.
- **Operations Staff**: manages scheduling, regional rules, analytics dashboards, exports, email queue, notifications.
- **Administrator**: full RBAC control, regional/service management, audits, reporting, catalog publishing.

## Navigation & Menu Visibility
| Section | Traveler | Coach/Clinician | Operations | Administrator |
| --- | --- | --- | --- | --- |
| Dashboard | ✓ | ✓ (agenda view) | ✓ | ✓ |
| Catalog (+ destinations/routes/hotels/attractions) | ✓ | ✓ (publish control) | ✓ | ✓ |
| Bookings & Holds | ✓ | ✓ (assigned sessions) | ✓ (scheduling ops) | ✓ |
| Profile & Contacts | ✓ | ✓ | ✓ | ✓ |
| Community | ✓ | ✓ | ✓ | ✓ |
| Notifications | ✓ | ✓ | ✓ | ✓ |
| Analytics + Scheduled Reports | (read only summary) | (agenda metrics) | ✓ | ✓ |
| Admin + RBAC |  |  |  | ✓ |

## Endpoint Permissions (examples)
| Endpoint | Permission | Travelers | Coaches/Clinicians | Ops | Admin |
| --- | --- | --- | --- | --- | --- |
| `GET /api/v1/profile/addresses` | `PermTravelerAddressRead` | ✓ |  |  |  |
| `POST /api/v1/profile/contacts` | `PermTravelerContactsAdd` | ✓ |  |  |  |
| `POST /api/v1/bookings/holds` | `PermTravelerBookingHold` | ✓ |  |  |  |
| `GET /api/v1/scheduling/hosts/:id/agenda` | `PermHostAgendaRead` |  | ✓ (own) | ✓ | ✓ |
| `GET /api/v1/admin/users` | `PermAdminUsersRead` |  |  | ✓ | ✓ |
| `POST /api/v1/reports/schedule` | `PermOpsReportsSchedule` |  |  | ✓ | ✓ |
| `POST /api/v1/community/reports` | `PermCommunityWrite` | ✓ | ✓ | ✓ | ✓ |
| `POST /api/v1/admin/roles/assign` | `PermAdminRoleAssign` |  |  |  | ✓ |

## Data Scope Notes
- Travelers are limited to their own profile, holds, bookings, notifications, and community interactions (ownership enforced via middleware and service-layer filtering).
- Coaches/Clinicians see their own assigned bookings and host/room agendas. They cannot audit other providers' private data unless explicitly granted via operations/admin roles.
- Operations staff are scoped to region/service ownership (e.g., `region_id = user.region_id`). Directory listings respect the regional assignment and blocked postal codes.
- Administrators bypass regional scoping but all privileged access is audited via `role_changes` and `permissions_audit` logs (see backend audit tables).
