package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestIPAllowlistDeniedAndBypass(t *testing.T) {
	e := echo.New()
	mw := IPAllowlist(IPAllowlistConfig{
		Allow:      []string{"127.0.0.1/32"},
		TrustProxy: true,
		BypassRoutes: map[string]struct{}{
			"/health": {},
		},
	})

	h := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]bool{"ok": true})
	})

	reqDenied := httptest.NewRequest(http.MethodGet, "/secure", nil)
	reqDenied.Header.Set("X-Forwarded-For", "203.0.113.11")
	reqDenied.RemoteAddr = "127.0.0.1:1234"
	recDenied := httptest.NewRecorder()
	ctxDenied := e.NewContext(reqDenied, recDenied)
	ctxDenied.SetPath("/secure")
	_ = h(ctxDenied)
	if recDenied.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recDenied.Code)
	}

	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqHealth.Header.Set("X-Forwarded-For", "203.0.113.11")
	reqHealth.RemoteAddr = "127.0.0.1:1234"
	recHealth := httptest.NewRecorder()
	ctxHealth := e.NewContext(reqHealth, recHealth)
	ctxHealth.SetPath("/health")
	_ = h(ctxHealth)
	if recHealth.Code != http.StatusOK {
		t.Fatalf("expected 200 bypass for /health, got %d", recHealth.Code)
	}
}
