package middleware

import (
	"fmt"
	"net/http"

	"meridian/backend/internal/api/response"

	"github.com/labstack/echo/v4"
)

const (
	PermAuthMe                 = "auth.me"
	PermTravelerAddressRead    = "traveler.address.read"
	PermTravelerAddressAdd     = "traveler.address.add"
	PermTravelerContactsRead   = "traveler.contacts.read"
	PermTravelerContactsAdd    = "traveler.contacts.add"
	PermTravelerContactsDelete = "traveler.contacts.delete"
	PermTravelerBookingHold    = "traveler.booking.hold"
	PermTravelerBookingList    = "traveler.booking.list"
	PermTravelerBookingHist    = "traveler.booking.history"
	PermAdminRoleAssign        = "admin.role.assign"
	PermAdminUsersRead         = "admin.users.read"
	PermAdminAuditsRead        = "admin.audits.read"
	PermHostAgendaRead         = "scheduling.host.agenda.read"
	PermRoomAgendaRead         = "scheduling.room.agenda.read"
	PermTravelerAddressDelete  = "traveler.address.delete"
	PermTravelerBookingCancel  = "traveler.booking.cancel"
	PermCatalogRoutesRead      = "catalog.routes.read"
	PermCatalogHotelsRead      = "catalog.hotels.read"
	PermCatalogAttractRead     = "catalog.attractions.read"
	PermSchedulingSlotsRead    = "scheduling.slots.read"
	PermSchedulingHostsRead    = "scheduling.hosts.read"
	PermTravelerBookingConfirm = "traveler.booking.confirm"
	PermCommunityRead          = "community.read"
	PermCommunityWrite         = "community.write"
	PermNotificationsRead      = "notifications.read"
	PermNotificationsWrite     = "notifications.write"
	PermOpsAnalyticsRead       = "analytics.read"
	PermOpsAnalyticsExport     = "analytics.export"
	PermOpsReportsSchedule     = "analytics.schedule"
	PermOpsEmailQueue          = "email.queue"
	PermAdminRegions           = "admin.regions"
)

var PermissionRoles = map[string][]string{
	PermAuthMe:                 {"traveler", "coach", "clinician", "operations", "admin"},
	PermTravelerAddressRead:    {"traveler"},
	PermTravelerAddressAdd:     {"traveler"},
	PermTravelerContactsRead:   {"traveler"},
	PermTravelerContactsAdd:    {"traveler"},
	PermTravelerContactsDelete: {"traveler"},
	PermTravelerBookingHold:    {"traveler"},
	PermTravelerBookingList:    {"traveler"},
	PermTravelerBookingHist:    {"traveler"},
	PermTravelerAddressDelete:  {"traveler"},
	PermTravelerBookingCancel:  {"traveler"},
	PermAdminRoleAssign:        {"operations", "admin"},
	PermAdminUsersRead:         {"operations", "admin"},
	PermAdminAuditsRead:        {"operations", "admin"},
	PermHostAgendaRead:         {"coach", "clinician", "operations", "admin"},
	PermRoomAgendaRead:         {"operations", "admin"},
	PermCatalogRoutesRead:      {"traveler", "operations", "admin"},
	PermCatalogHotelsRead:      {"traveler", "operations", "admin"},
	PermCatalogAttractRead:     {"traveler", "operations", "admin"},
	PermSchedulingSlotsRead:    {"traveler", "operations", "admin"},
	PermSchedulingHostsRead:    {"traveler", "operations", "admin"},
	PermTravelerBookingConfirm: {"traveler", "admin"},
	PermCommunityRead:          {"traveler", "coach", "clinician", "operations", "admin"},
	PermCommunityWrite:         {"traveler", "coach", "clinician", "operations", "admin"},
	PermNotificationsRead:      {"traveler", "coach", "clinician", "operations", "admin"},
	PermNotificationsWrite:     {"traveler", "coach", "clinician", "operations", "admin"},
	PermOpsAnalyticsRead:       {"operations", "admin"},
	PermOpsAnalyticsExport:     {"operations", "admin"},
	PermOpsReportsSchedule:     {"operations", "admin"},
	PermOpsEmailQueue:          {"operations", "admin"},
	PermAdminRegions:           {"operations", "admin"},
}

func RequirePermission(permission string) echo.MiddlewareFunc {
	allowed := map[string]struct{}{}
	for _, role := range PermissionRoles[permission] {
		allowed[role] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles := RolesFromContext(c)
			for _, role := range roles {
				if _, ok := allowed[role]; ok {
					return next(c)
				}
			}
			uid, _ := UserID(c)
			c.Logger().Warn(fmt.Sprintf("permission denied user=%d permission=%s roles=%v", uid, permission, roles))
			return response.JSONError(c, http.StatusForbidden, "insufficient permission")
		}
	}
}

func RolesFromContext(c echo.Context) []string {
	rolesAny := c.Get("roles")
	raw, ok := rolesAny.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		if rs, ok := r.(string); ok {
			out = append(out, rs)
		}
	}
	return out
}

func HasAnyRole(c echo.Context, roles ...string) bool {
	want := map[string]struct{}{}
	for _, role := range roles {
		want[role] = struct{}{}
	}
	for _, actual := range RolesFromContext(c) {
		if _, ok := want[actual]; ok {
			return true
		}
	}
	return false
}
