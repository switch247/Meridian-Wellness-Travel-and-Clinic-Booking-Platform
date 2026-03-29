package middleware

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLoggerRedactsAuth(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	mw := RequestLogger(logger)
	dummy := func(c echo.Context) error { return nil }
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/secure")

	h := mw(dummy)
	_ = h(ctx)

	logOut := buf.String()
	if strings.Contains(logOut, "secret-token") {
		t.Fatalf("log output leaked token: %s", logOut)
	}
}
