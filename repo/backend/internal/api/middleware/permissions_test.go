package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRolesFromContextAndHasAnyRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("roles", []any{"traveler", "coach"})

	roles := RolesFromContext(c)
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if !HasAnyRole(c, "coach") {
		t.Fatalf("expected HasAnyRole to find coach")
	}
	if HasAnyRole(c, "admin") {
		t.Fatalf("did not expect admin role")
	}
}

func TestRequirePermissionDenyAndAllow(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// traveler should be denied admin users permission
	c.Set("roles", []any{"traveler"})
	c.Set("userID", int64(1))

	handler := RequirePermission(PermAdminUsersRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for traveler, got %d", rec.Code)
	}

	// admin should be allowed
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req, rec2)
	c2.Set("roles", []any{"admin"})
	c2.Set("userID", int64(2))
	handler2 := RequirePermission(PermAdminUsersRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	_ = handler2(c2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", rec2.Code)
	}
}
