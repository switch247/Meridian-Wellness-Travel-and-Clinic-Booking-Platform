package security

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	upperRe = regexp.MustCompile(`[A-Z]`)
	lowerRe = regexp.MustCompile(`[a-z]`)
	digitRe = regexp.MustCompile(`[0-9]`)
	symRe   = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	if !upperRe.MatchString(password) || !lowerRe.MatchString(password) || !digitRe.MatchString(password) || !symRe.MatchString(password) {
		return errors.New("password must include uppercase, lowercase, number, and symbol")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ComparePassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
