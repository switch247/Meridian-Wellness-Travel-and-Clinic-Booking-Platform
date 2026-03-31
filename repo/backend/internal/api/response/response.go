package response

import (
	"net/http"

	internalLogger "meridian/backend/internal/logger"

	"github.com/labstack/echo/v4"
)

type ErrorBody struct {
	Error string `json:"error"`
}

func JSONError(c echo.Context, status int, msg string) error {
	// For server errors, do not expose internal messages to clients.
	if status >= http.StatusInternalServerError {
		// Log the internal message with structured logger (redactions applied by logger package).
		l := internalLogger.New()
		l.Error("internal_error", "message", msg, "path", c.Request().URL.Path)
		return c.JSON(status, ErrorBody{Error: "internal server error"})
	}
	return c.JSON(status, ErrorBody{Error: msg})
}

func BindAndValidate[T any](c echo.Context, dst *T, validator func(*T) error) error {
	if err := c.Bind(dst); err != nil {
		return JSONError(c, http.StatusBadRequest, "invalid payload")
	}
	if err := validator(dst); err != nil {
		return JSONError(c, http.StatusBadRequest, err.Error())
	}
	return nil
}
