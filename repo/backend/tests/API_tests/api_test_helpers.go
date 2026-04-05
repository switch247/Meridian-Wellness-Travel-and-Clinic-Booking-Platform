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
	res, body := call(http.MethodPost, "/auth/login", "", map[string]any{"username": "admin", "password": "Password123!"}, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin login failed %d %+v", res.StatusCode, body)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatal("missing admin token")
	}
	return tok
}
