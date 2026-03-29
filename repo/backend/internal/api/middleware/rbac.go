package middleware

import (
	"net/http"

	"meridian/backend/internal/api/response"

	"github.com/labstack/echo/v4"
)

func RequireRole(allowed ...string) echo.MiddlewareFunc {
	allowedSet := map[string]struct{}{}
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rolesAny := c.Get("roles")
			rolesSlice, ok := rolesAny.([]any)
			if !ok {
				return response.JSONError(c, http.StatusForbidden, "no roles in token")
			}
			for _, r := range rolesSlice {
				rs, ok := r.(string)
				if ok {
					if _, exists := allowedSet[rs]; exists {
						return next(c)
					}
				}
			}
			return response.JSONError(c, http.StatusForbidden, "insufficient role")
		}
	}
}
