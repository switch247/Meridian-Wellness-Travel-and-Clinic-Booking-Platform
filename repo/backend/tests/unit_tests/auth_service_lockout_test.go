package unit_tests

import (
	"context"
	"os"
	"testing"
	"time"

	"log/slog"

	"meridian/backend/internal/config"
	"meridian/backend/internal/platform/db"
	"meridian/backend/internal/repository"
	"meridian/backend/internal/security"
	"meridian/backend/internal/service"

	"github.com/google/uuid"
)

func TestAuthService_LockoutAfterFailedAttempts(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping DB integration test")
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	repo := repository.New(pool)
	enc, err := security.NewEncryptor("01234567890123456789012345678901")
	if err != nil {
		t.Fatalf("new encryptor: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Config{
		LockoutThreshold: 3,
		LockoutDuration:  2 * time.Minute,
		JWTSecret:        "test-jwt-secret",
		TokenTTL:         1 * time.Hour,
	}
	svc := service.NewAuthService(repo, cfg, enc, logger)

	username := "lockout_" + uuid.NewString()
	password := "Str0ngP@ssw0rd!"
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	uid, err := repo.CreateUser(ctx, username, hash, "", "")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Attempt to login with wrong password LockoutThreshold times
	for i := 0; i < cfg.LockoutThreshold; i++ {
		_, err := svc.Login(ctx, username, "wrong-password")
		if err == nil {
			t.Fatalf("expected login to fail on attempt %d", i+1)
		}
	}

	// Verify user is locked
	u, _, err := repo.FindUserByUsername(ctx, username)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if u.LockedUntil == nil {
		t.Fatalf("expected user to be locked, but LockedUntil is nil")
	}
	if !u.LockedUntil.After(time.Now()) {
		t.Fatalf("expected LockedUntil in future, got %v", u.LockedUntil)
	}

	// Clean up: reset failed attempts
	if err := repo.ResetFailedAttempts(ctx, uid); err != nil {
		t.Fatalf("reset attempts: %v", err)
	}
}
