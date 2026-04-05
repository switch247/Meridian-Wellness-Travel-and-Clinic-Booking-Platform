# Role Matrix & Visibility

This matrix captures the high-level data and endpoint surface each built-in role can reach, as enforced by `middleware.RequirePermission` in the backend router.

| Role | Key Permissions | Notes |
| --- | --- | --- |
| `admin` | `/admin/*` + RBAC assignment, region/site configuration | Super-admin access, can read/write user/permission audits, publish catalog entries, manage service rules. |
| `operations` | `/ops/*`, `/scheduling/*`, `/community/*` (augmented) | Reads analytics KPIs, schedule data, and books/cancels sessions across the tenant. |
| `coach` / `clinician` | `/scheduling/hosts/:id/agenda`, `/community/*` | Granted host agenda, notification, community write/read rights scoped to their tenant plus ownership guards. |
| `traveler` | `/bookings/*`, `/profile/*`, `/community/*` | Can place holds, confirm bookings, manage their profile/contact/address book, and interact in community threads. |

## Enforcement
- **Request guards** (`middleware.RequirePermission`) gate every route shown above.  
- **Domain handler middleware** (`JWT`, `LocationID`) also injects `locationId` and ownership checks per request.  
- **Repository checks** enforce `location_id` filters and block lists so the request trace aligns with the permission matrix above.
