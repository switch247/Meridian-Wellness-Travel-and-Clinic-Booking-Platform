package repository_test

import (
	"context"
	"testing"
	"time"

	"meridian/backend/internal/repository"
	"meridian/backend/tests/testutil"

	"github.com/google/uuid"
)

func TestConfirmHold_ExpiredRejected(t *testing.T) {
	pool := testutil.DBPoolOrSkip(t)
	defer pool.Close()
	ctx := context.Background()
	repo := repository.New(pool)

	// Create a temporary user
	username := "testuser_" + uuid.NewString()
	uid, err := repo.CreateUser(ctx, username, "hash", "", "")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create a minimal destination and package to satisfy FK
	var destID int64
	if err := pool.QueryRow(ctx, `INSERT INTO destinations(name,description,image_path) VALUES($1,$2,$3) RETURNING id`, "d", "d", "p").Scan(&destID); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	var pkgID int64
	if err := pool.QueryRow(ctx, `INSERT INTO packages(destination_id,name,description) VALUES($1,$2,$3) RETURNING id`, destID, "p", "p").Scan(&pkgID); err != nil {
		t.Fatalf("create package: %v", err)
	}

	// Insert an expired hold
	var holdID int64
	slot := time.Now().UTC().Add(1 * time.Hour)
	expires := time.Now().UTC().Add(-1 * time.Minute)
	if err := pool.QueryRow(ctx, `INSERT INTO reservation_holds(user_id,package_id,host_id,room_id,slot_start,duration_minutes,expires_at,status,version) VALUES($1,$2,$3,$4,$5,$6,$7,'active',1) RETURNING id`, uid, pkgID, 1, 1, slot, 45, expires).Scan(&holdID); err != nil {
		t.Fatalf("insert hold: %v", err)
	}

	// Attempt to confirm; expect ErrHoldExpired
	_, _, err = repo.ConfirmHold(ctx, uid, holdID, 1)
	if err == nil {
		t.Fatalf("expected error confirming expired hold")
	}
	if err != repository.ErrHoldExpired {
		t.Fatalf("expected ErrHoldExpired, got %v", err)
	}
}
