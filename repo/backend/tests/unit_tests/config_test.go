package unit_tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"meridian/backend/internal/api/handlers"
	"meridian/backend/internal/config"

	"github.com/labstack/echo/v4"
)

func TestConfigCoverageReturnsAllowedRegions(t *testing.T) {
	cfg := config.Config{AllowedRegions: []string{"10001", "60601"}}
	domain := handlers.NewDomainHandler(nil, nil, nil, nil, cfg, 15)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/config/coverage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := domain.ConfigCoverage(c); err != nil {
		t.Fatalf("config handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	var payload map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := payload["allowedRegions"]; len(got) != 2 || got[0] != "10001" || got[1] != "60601" {
		t.Fatalf("unexpected regions payload: %v", payload)
	}
}
