package security

import "testing"

func TestNewEncryptorAndRoundtrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef" // 32 bytes
	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("unexpected error creating encryptor: %v", err)
	}
	plain := "sensitive notes"
	out, err := enc.Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	dec, err := enc.Decrypt(out)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if dec != plain {
		t.Fatalf("decrypt mismatch: got %q want %q", dec, plain)
	}
}

func TestNewEncryptorBadKey(t *testing.T) {
	_, err := NewEncryptor("shortkey")
	if err == nil {
		t.Fatalf("expected error for short key")
	}
}
