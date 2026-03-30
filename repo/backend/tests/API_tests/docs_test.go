package integration_tests

import (
	"os"
	"testing"
)

func TestDocsEndpoints(t *testing.T) {
	if os.Getenv("BASE_URL") == "" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("integration tests skipped; set RUN_INTEGRATION_TESTS=true or BASE_URL to run")
	}
	res, err := testClient().Get(baseURL() + "/docs")
	if err != nil {
		t.Fatalf("request docs: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("expected 200 from docs landing, got %d", res.StatusCode)
	}

	res2, err := testClient().Get(baseURL() + "/docs/openapi.yaml")
	if err != nil {
		t.Fatalf("request openapi: %v", err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != 200 {
		t.Fatalf("expected 200 from openapi, got %d", res2.StatusCode)
	}
}
