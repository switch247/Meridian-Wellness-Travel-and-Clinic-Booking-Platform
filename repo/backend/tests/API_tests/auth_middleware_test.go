package integration_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"meridian/backend/internal/api/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const testJWTSecret = "test-secret"

func makeToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func executeWithJWT(t *testing.T, authHeader string, next echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware.JWT(middleware.AuthConfig{JWTSecret: testJWTSecret})(next)
	_ = handler(c)
	return rec
}

func TestJWTMiddlewareRejectsMissingToken(t *testing.T) {
	rec := executeWithJWT(t, "", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rec.Code)
	}
}

func TestJWTMiddlewareRejectsWrongSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{"sub": 1, "roles": []string{"traveler"}})
	raw, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	rec := executeWithJWT(t, "Bearer "+raw, func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rec.Code)
	}
}

func TestJWTMiddlewareSetsUserAndLocationContext(t *testing.T) {
	raw := makeToken(t, jwt.MapClaims{
		"sub":        77,
		"roles":      []string{"coach"},
		"locationId": 12,
	})

	rec := executeWithJWT(t, "Bearer "+raw, func(c echo.Context) error {
		userID, ok := middleware.UserID(c)
		if !ok || userID != 77 {
			t.Fatalf("expected userID=77, got %d, ok=%v", userID, ok)
		}
		locationID, ok := middleware.LocationID(c)
		if !ok || locationID != 12 {
			t.Fatalf("expected locationID=12, got %d, ok=%v", locationID, ok)
		}
		return c.NoContent(http.StatusOK)
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}
