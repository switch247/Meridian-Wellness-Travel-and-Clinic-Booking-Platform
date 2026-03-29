package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorBody struct {
	Error string `json:"error"`
}

func JSONError(c echo.Context, status int, msg string) error {
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
