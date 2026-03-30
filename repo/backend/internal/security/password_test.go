package security

import "testing"

func TestValidatePasswordSuccess(t *testing.T) {
	if err := ValidatePassword("Strong#Pass123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePasswordFailures(t *testing.T) {
	cases := []string{
		"short1A#",
		"nouppercase123#",
		"NOLOWERCASE123#",
		"NoNumberOrSymbol",
	}
	for _, c := range cases {
		if err := ValidatePassword(c); err == nil {
			t.Fatalf("expected error for password: %s", c)
		}
	}
}
