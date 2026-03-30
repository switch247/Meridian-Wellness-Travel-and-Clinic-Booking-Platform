package unit_tests

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
		return "https://localhost:8443/api/v1"
	}
	return v + "/api/v1"
}

func testClient() *http.Client {
	if os.Getenv("ALLOW_INSECURE_TLS") == "true" {
		return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}
	return http.DefaultClient
}

func postJSON(t *testing.T, path string, payload any, token string) (*http.Response, map[string]any) {
	t.Helper()
	if os.Getenv("BASE_URL") == "" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("integration tests skipped; set RUN_INTEGRATION_TESTS=true or BASE_URL to run")
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, baseURL()+path, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := testClient().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	return res, body
}

func getJSON(t *testing.T, path string, token string) (*http.Response, map[string]any) {
	t.Helper()
	if os.Getenv("BASE_URL") == "" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("integration tests skipped; set RUN_INTEGRATION_TESTS=true or BASE_URL to run")
	}
	req, err := http.NewRequest(http.MethodGet, baseURL()+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := testClient().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	return res, body
}

func loginToken(t *testing.T) string {
	t.Helper()
	u := fmt.Sprintf("u_%d", time.Now().UnixNano())
	_, _ = postJSON(t, "/auth/register", map[string]any{
		"username": u,
		"password": "Strong#Pass123",
		"phone":    "+15550001111",
		"address":  "100 Main Street New York",
	}, "")
	res, body := postJSON(t, "/auth/login", map[string]any{"username": u, "password": "Strong#Pass123"}, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %d %+v", res.StatusCode, body)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatal("missing token")
	}
	return tok
}

func TestPasswordPolicyBoundary(t *testing.T) {
	u := fmt.Sprintf("pw_%d", time.Now().UnixNano())
	res, _ := postJSON(t, "/auth/register", map[string]any{
		"username": u,
		"password": "weakpassword",
		"phone":    "+15550002222",
		"address":  "200 Main Street New York",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for weak password, got %d", res.StatusCode)
	}
}

func TestAddressNormalizationAndCoverage(t *testing.T) {
	token := loginToken(t)
	res, body := postJSON(t, "/profile/addresses", map[string]any{
		"line1":      "123 Main Street",
		"line2":      "",
		"city":       "New York",
		"state":      "NY",
		"postalCode": "99999",
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%+v", res.StatusCode, body)
	}
	if body["inCoverage"] != false {
		t.Fatalf("expected out-of-coverage warning")
	}

	resList, bodyList := getJSON(t, "/profile/addresses", token)
	if resList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for address list, got %d", resList.StatusCode)
	}
	if _, ok := bodyList["items"]; !ok {
		t.Fatalf("expected items key")
	}
}
