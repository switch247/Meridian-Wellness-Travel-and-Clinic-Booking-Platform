package api

import (
	"log/slog"
	"net/http"

	"meridian/backend/internal/api/handlers"
	"meridian/backend/internal/api/middleware"
	"meridian/backend/internal/config"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func NewRouter(cfg config.Config, logger *slog.Logger, authH *handlers.AuthHandler, domainH *handlers.DomainHandler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger(logger))
	e.Use(middleware.SecurityHeaders())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	e.Use(middleware.IPAllowlist(middleware.IPAllowlistConfig{
		Allow:      cfg.AllowedIPs,
		TrustProxy: cfg.TrustProxyHeaders,
		BypassRoutes: map[string]struct{}{
			"/health": {},
		},
	}))

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		logger.Error("request_error", "path", c.Path(), "method", c.Request().Method, "error", err.Error())
		_ = c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.File("/docs/openapi.yaml", "docs/openapi.yaml")
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	e.GET("/docs/*", echoSwagger.EchoWrapHandler(echoSwagger.URL("/docs/openapi.yaml")))

	v1 := e.Group("/api/v1")
	v1.POST("/auth/register", authH.Register)
	v1.POST("/auth/login", authH.Login)
	v1.GET("/catalog", domainH.Catalog)
	v1.GET("/catalog/routes", domainH.Routes)
	v1.GET("/catalog/hotels", domainH.Hotels)
	v1.GET("/catalog/attractions", domainH.Attractions)

	authed := v1.Group("", middleware.JWT(middleware.AuthConfig{JWTSecret: cfg.JWTSecret}))
	authed.GET("/auth/me", authH.Me, middleware.RequirePermission(middleware.PermAuthMe))

	authed.POST("/profile/addresses", domainH.AddAddress, middleware.RequirePermission(middleware.PermTravelerAddressAdd))
	authed.GET("/profile/addresses", domainH.ListAddresses, middleware.RequirePermission(middleware.PermTravelerAddressRead))
	authed.POST("/profile/contacts", domainH.AddContact, middleware.RequirePermission(middleware.PermTravelerContactsAdd))
	authed.GET("/profile/contacts", domainH.ListContacts, middleware.RequirePermission(middleware.PermTravelerContactsRead))

	authed.GET("/users/:id", domainH.GetUser, middleware.RequirePermission(middleware.PermAuthMe))
	authed.POST("/bookings/holds", domainH.PlaceHold, middleware.RequirePermission(middleware.PermTravelerBookingHold))
	authed.GET("/bookings/holds", domainH.ListHolds, middleware.RequirePermission(middleware.PermTravelerBookingList))
	authed.GET("/bookings/history", domainH.ListBookingHistory, middleware.RequirePermission(middleware.PermTravelerBookingHist))
	authed.POST("/bookings/confirm", domainH.ConfirmHold, middleware.RequirePermission(middleware.PermTravelerBookingConfirm))
	authed.POST("/bookings/:id/status", domainH.UpdateBookingStatus, middleware.RequirePermission(middleware.PermBookingStatusUpdate))
	authed.GET("/scheduling/slots", domainH.AvailableSlots, middleware.RequirePermission(middleware.PermSchedulingSlotsRead))
	authed.GET("/scheduling/hosts", domainH.ListHosts, middleware.RequirePermission(middleware.PermSchedulingHostsRead))
	authed.GET("/scheduling/rooms", domainH.ListRooms, middleware.RequirePermission(middleware.PermSchedulingHostsRead))

	authed.GET("/scheduling/hosts/:id/agenda", domainH.HostAgenda, middleware.RequirePermission(middleware.PermHostAgendaRead))
	authed.GET("/scheduling/rooms/:id/agenda", domainH.RoomAgenda, middleware.RequirePermission(middleware.PermRoomAgendaRead))
	authed.DELETE("/profile/addresses/:id", domainH.DeleteAddress, middleware.RequirePermission(middleware.PermTravelerAddressDelete))
	authed.DELETE("/profile/contacts/:id", domainH.DeleteContact, middleware.RequirePermission(middleware.PermTravelerContactsDelete))
	authed.DELETE("/bookings/holds/:id", domainH.CancelHold, middleware.RequirePermission(middleware.PermTravelerBookingCancel))

	admin := authed.Group("/admin")
	admin.POST("/roles/assign", domainH.AssignRole, middleware.RequirePermission(middleware.PermAdminRoleAssign))
	admin.GET("/users", domainH.ListUsers, middleware.RequirePermission(middleware.PermAdminUsersRead))
	admin.GET("/roles/audits", domainH.ListRoleAudits, middleware.RequirePermission(middleware.PermAdminAuditsRead))
	admin.POST("/reports/:id/resolve", domainH.ResolveReport, middleware.RequirePermission(middleware.PermAdminAuditsRead))
	admin.GET("/regions", domainH.ListRegions, middleware.RequirePermission(middleware.PermAdminRegions))
	admin.POST("/regions", domainH.CreateRegion, middleware.RequirePermission(middleware.PermAdminRegions))
	admin.POST("/regions/:id/service-rule", domainH.UpsertServiceRule, middleware.RequirePermission(middleware.PermAdminRegions))
	admin.GET("/service/blocked-postal-codes", domainH.ListBlockedPostalCodes, middleware.RequirePermission(middleware.PermAdminRegions))
	admin.POST("/service/blocked-postal-codes", domainH.AddBlockedPostalCode, middleware.RequirePermission(middleware.PermAdminRegions))

	community := authed.Group("/community")
	community.GET("/posts", domainH.ListPosts, middleware.RequirePermission(middleware.PermCommunityRead))
	community.POST("/posts", domainH.CreatePost, middleware.RequirePermission(middleware.PermCommunityWrite))
	community.GET("/posts/:id/comments", domainH.ListComments, middleware.RequirePermission(middleware.PermCommunityRead))
	community.POST("/posts/:id/comments", domainH.CreateComment, middleware.RequirePermission(middleware.PermCommunityWrite))
	community.POST("/favorites", domainH.FavoritePackage, middleware.RequirePermission(middleware.PermCommunityWrite))
	community.POST("/likes", domainH.LikeTarget, middleware.RequirePermission(middleware.PermCommunityWrite))
	community.POST("/follows", domainH.FollowUser, middleware.RequirePermission(middleware.PermCommunityWrite))
	community.POST("/blocks", domainH.BlockUser, middleware.RequirePermission(middleware.PermCommunityWrite))
	community.POST("/reports", domainH.ReportContent, middleware.RequirePermission(middleware.PermCommunityWrite))

	authed.GET("/notifications", domainH.ListNotifications, middleware.RequirePermission(middleware.PermNotificationsRead))
	authed.POST("/notifications/:id/read", domainH.MarkNotificationRead, middleware.RequirePermission(middleware.PermNotificationsWrite))

	ops := authed.Group("/ops")
	ops.GET("/analytics/kpis", domainH.AnalyticsKPIs, middleware.RequirePermission(middleware.PermOpsAnalyticsRead))
	ops.GET("/analytics/export", domainH.ExportAnalytics, middleware.RequirePermission(middleware.PermOpsAnalyticsExport))
	ops.GET("/reports", domainH.ListReportJobs, middleware.RequirePermission(middleware.PermOpsAnalyticsRead))
	ops.POST("/reports/schedule", domainH.ScheduleReport, middleware.RequirePermission(middleware.PermOpsReportsSchedule))
	ops.POST("/email/queue", domainH.QueueEmail, middleware.RequirePermission(middleware.PermOpsEmailQueue))
	ops.GET("/email/queue", domainH.ListEmailQueue, middleware.RequirePermission(middleware.PermOpsEmailQueue))
	ops.POST("/email/export", domainH.ExportEmailQueue, middleware.RequirePermission(middleware.PermOpsEmailQueue))

	return e
}
