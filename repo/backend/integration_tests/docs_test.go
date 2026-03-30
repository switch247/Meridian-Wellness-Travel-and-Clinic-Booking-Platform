package integration_tests

import "testing"

func TestDocsEndpoints(t *testing.T) {
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
