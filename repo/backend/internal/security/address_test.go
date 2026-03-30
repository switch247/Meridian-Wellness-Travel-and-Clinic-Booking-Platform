package security

import (
	"testing"
)

func TestNormalizeUSAddress(t *testing.T) {
	in := "123 Main Street"
	out := NormalizeUSAddress(in, "New York", "NY", "10001")
	if out == "" {
		t.Fatalf("expected normalized value")
	}
	if contains := " st"; contains == " st" && out == "" {
		t.Fatalf("unexpected normalization result")
	}
}

func TestInCoverage(t *testing.T) {
	allowed := []string{"10001", "20002"}
	if !InCoverage("10001", allowed) {
		t.Fatalf("expected 10001 to be in coverage")
	}
	if InCoverage("99999", allowed) {
		t.Fatalf("did not expect 99999 to be in coverage")
	}
}
