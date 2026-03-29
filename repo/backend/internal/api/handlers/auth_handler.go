package handlers

import (
	"net/http"

	"meridian/backend/internal/api/middleware"
	"meridian/backend/internal/api/response"
	"meridian/backend/internal/service"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := response.BindAndValidate(c, &req, func(r *registerRequest) error {
		if r.Username == "" || r.Password == "" || r.Phone == "" || r.Address == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username, password, phone, address required")
		}
		return nil
	}); err != nil {
		return err
	}
	id, err := h.auth.Register(c.Request().Context(), req.Username, req.Password, req.Phone, req.Address)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := response.BindAndValidate(c, &req, func(r *loginRequest) error {
		if r.Username == "" || r.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username and password required")
		}
		return nil
	}); err != nil {
		return err
	}
	result, err := h.auth.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return response.JSONError(c, http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	payload, err := h.auth.Me(c.Request().Context(), userID)
	if err != nil {
		return response.JSONError(c, http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, payload)
}
