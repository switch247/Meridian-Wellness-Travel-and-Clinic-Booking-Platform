package integration_tests

import (
	"net/http"
	"testing"
)

func TestCatalogHotelsFunctional(t *testing.T) {
	res, body := call(http.MethodGet, "/catalog/hotels", "", nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /catalog/hotels, got %d body=%+v", res.StatusCode, body)
	}
	if _, ok := body["items"]; !ok {
		t.Fatalf("expected items field in /catalog/hotels response")
	}
}

func TestCatalogAttractionsFunctional(t *testing.T) {
	res, body := call(http.MethodGet, "/catalog/attractions", "", nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /catalog/attractions, got %d body=%+v", res.StatusCode, body)
	}
	if _, ok := body["items"]; !ok {
		t.Fatalf("expected items field in /catalog/attractions response")
	}
}
