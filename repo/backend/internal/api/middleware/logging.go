package middleware

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.NewString()
			}
			c.Response().Header().Set("X-Request-ID", reqID)
			c.Set("requestID", reqID)
			return next(c)
		}
	}
}

// Redacts auth headers and logs structured request summary.
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			status := c.Response().Status
			req := c.Request()
			reqID, _ := c.Get("requestID").(string)
			ip := req.RemoteAddr
			logger.Info("request",
				"req_id", reqID,
				"method", req.Method,
				"path", c.Path(),
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", ip,
			)
			return err
		}
	}
}
