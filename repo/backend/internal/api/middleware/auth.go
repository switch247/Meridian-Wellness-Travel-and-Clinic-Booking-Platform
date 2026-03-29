package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"meridian/backend/internal/api/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthConfig struct {
	JWTSecret string
}

func JWT(cfg AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tok := c.Request().Header.Get("Authorization")
			if tok == "" {
				return response.JSONError(c, http.StatusUnauthorized, "missing bearer token")
			}
			const p = "Bearer "
			if len(tok) <= len(p) || tok[:len(p)] != p {
				return response.JSONError(c, http.StatusUnauthorized, "invalid authorization header")
			}
	parsed, err := jwt.Parse(tok[len(p):], func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return response.JSONError(c, http.StatusUnauthorized, "invalid token")
	}
			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				return response.JSONError(c, http.StatusUnauthorized, "invalid token claims")
			}
			sub, ok := claims["sub"]
			if !ok {
				return response.JSONError(c, http.StatusUnauthorized, "missing subject")
			}

			var userID int64
			switch v := sub.(type) {
			case float64:
				userID = int64(v)
			case string:
				n, convErr := strconv.ParseInt(v, 10, 64)
				if convErr != nil {
					return response.JSONError(c, http.StatusUnauthorized, "invalid subject")
				}
				userID = n
			default:
				return response.JSONError(c, http.StatusUnauthorized, "invalid subject type")
			}
			c.Set("userID", userID)
			c.Set("roles", claims["roles"])
			return next(c)
		}
	}
}

func UserID(c echo.Context) (int64, bool) {
	v := c.Get("userID")
	id, ok := v.(int64)
	return id, ok
}
