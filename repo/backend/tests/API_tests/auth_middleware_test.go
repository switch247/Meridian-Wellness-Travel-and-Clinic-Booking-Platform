package integration_tests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func baseURL() string {
	v := os.Getenv("BASE_URL")
	if v == "" {
		return "https://localhost:8443"
	}
	return v
}

func apiBase() string {
	return baseURL() + "/api/v1"
}

func testClient() *http.Client {
	insecure := os.Getenv("ALLOW_INSECURE_TLS") == "true"
	if insecure {
		return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}
	return http.DefaultClient
}

func call(method, path, token string, payload any, t *testing.T) (*http.Response, map[string]any) {
	t.Helper()
	if os.Getenv("BASE_URL") == "" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("integration tests skipped; set RUN_INTEGRATION_TESTS=true or BASE_URL to run")
	}
	var bodyBytes []byte
	if payload != nil {
		bodyBytes, _ = json.Marshal(payload)
	}
	req, err := http.NewRequest(method, apiBase()+path, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := testClient().Do(req)
	if err != nil {
		t.Fatalf("http call: %v", err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res, out
}

func makeUserToken(t *testing.T) string {
	t.Helper()
	u := fmt.Sprintf("api_%d", time.Now().UnixNano())
	_, _ = call(http.MethodPost, "/auth/register", "", map[string]any{
		"username": u,
		"password": "Strong#Pass123",
		"phone":    "+15558889999",
		"address":  "12 Avenue Road New York",
	}, t)
	res, body := call(http.MethodPost, "/auth/login", "", map[string]any{"username": u, "password": "Strong#Pass123"}, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login failed %d %+v", res.StatusCode, body)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatal("empty token")
	}
	return tok
}

func makeAdminToken(t *testing.T) string {
	t.Helper()
	res, body := call(http.MethodPost, "/auth/login", "", map[string]any{"username": "admin", "password": "Admin#Pass123"}, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin login failed %d %+v", res.StatusCode, body)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatal("missing admin token")
	}
	return tok
}

func TestHTTPSAndAuthGuards(t *testing.T) {
	res, _ := call(http.MethodGet, "/auth/me", "", nil, t)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", res.StatusCode)
	}
}

func TestForbiddenOnAdminRouteForTraveler(t *testing.T) {
	token := makeUserToken(t)
	res, _ := call(http.MethodPost, "/admin/roles/assign", token, map[string]any{
		"targetUserId": 1,
		"role":         "operations",
	}, t)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", res.StatusCode)
	}
}

func TestHTTPIsNotAvailableWhenTLSEnabled(t *testing.T) {
	res, err := http.Get("http://localhost:8443/health")
	if err == nil && res != nil && res.StatusCode == http.StatusOK {
		t.Fatalf("expected plain HTTP to not succeed with 200 on TLS port")
	}
}

func TestTravelerFlowEndpoints(t *testing.T) {
	token := makeUserToken(t)
	resAddr, _ := call(http.MethodGet, "/profile/addresses", token, nil, t)
	if resAddr.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for address list, got %d", resAddr.StatusCode)
	}
	resHolds, _ := call(http.MethodGet, "/bookings/holds", token, nil, t)
	if resHolds.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for holds list, got %d", resHolds.StatusCode)
	}
	resHist, bodyHist := call(http.MethodGet, "/bookings/history", token, nil, t)
	if resHist.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for booking history, got %d", resHist.StatusCode)
	}
	if _, ok := bodyHist["items"]; !ok {
		t.Fatalf("expected items key in history response")
	}
}

func TestAdminAndStaffEndpoints(t *testing.T) {
	token := makeAdminToken(t)

	resUsers, _ := call(http.MethodGet, "/admin/users", token, nil, t)
	if resUsers.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for admin users list, got %d", resUsers.StatusCode)
	}

	resAudits, _ := call(http.MethodGet, "/admin/roles/audits", token, nil, t)
	if resAudits.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for role audits, got %d", resAudits.StatusCode)
	}

	resRoom, _ := call(http.MethodGet, "/scheduling/rooms/1/agenda", token, nil, t)
	if resRoom.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for room agenda, got %d", resRoom.StatusCode)
	}
}
