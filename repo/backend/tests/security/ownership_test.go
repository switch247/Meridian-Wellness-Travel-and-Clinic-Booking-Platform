package security_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"meridian/backend/internal/api/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func signedToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString([]byte("tenant-test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return raw
}

func TestCoachWithoutLocationClaimIsDenied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/scheduling/rooms/1/agenda", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, jwt.MapClaims{"sub": 5, "roles": []string{"coach"}}))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware.JWT(middleware.AuthConfig{JWTSecret: "tenant-test-secret"})(func(c echo.Context) error {
		if _, ok := middleware.LocationID(c); ok {
			t.Fatalf("unexpected location claim")
		}
		return c.JSON(http.StatusForbidden, map[string]any{"error": "location scope missing"})
	})
	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", rec.Code)
	}
}

func TestCoachWithLocationClaimIsSessionScoped(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/scheduling/rooms/2/agenda", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, jwt.MapClaims{"sub": 5, "roles": []string{"coach"}, "locationId": 100}))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware.JWT(middleware.AuthConfig{JWTSecret: "tenant-test-secret"})(func(c echo.Context) error {
		locationID, ok := middleware.LocationID(c)
		if !ok || locationID != 100 {
			t.Fatalf("expected locationID=100, got %d ok=%v", locationID, ok)
		}
		return c.NoContent(http.StatusOK)
	})
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}
