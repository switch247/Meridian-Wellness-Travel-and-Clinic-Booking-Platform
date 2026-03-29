package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRedactSensitive(t *testing.T) {
	if got := RedactSensitive("1234567890"); got == "1234567890" {
		t.Fatalf("expected masked value")
	}
}

func TestRotatingWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w, err := NewRotatingWriter(path, 64, 2)
	if err != nil {
		t.Fatalf("new rotating writer: %v", err)
	}
	defer w.Close()
	for i := 0; i < 20; i++ {
		if _, err := w.Write([]byte("0123456789abcdef\n")); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("expected rotated file: %v", err)
	}
}

func TestLoggerRedactsAuthAttr(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithWriter(&buf)
	l.Info("auth", "authorization", "Bearer abc", "phone", "+15551234567")
	out := buf.String()
	if out == "" {
		t.Fatalf("expected logs")
	}
	if contains(out, "Bearer abc") {
		t.Fatalf("authorization not redacted")
	}
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
