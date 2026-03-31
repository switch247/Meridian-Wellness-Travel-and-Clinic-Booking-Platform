package repository

import (
	"context"
	"testing"
	"time"

	"meridian/backend/tests/testutil"
)

func TestChairHoldAndConfirmConflicts(t *testing.T) {
	pool := testutil.DBPoolOrSkip(t)
	repo := New(pool)
	ctx := context.Background()

	// Use an existing room from seed or create a temp one
	// For simplicity assume room id 1 exists (seed creates Room A)
	roomID := int64(1)

	// create a chair
	chairID, err := repo.CreateChair(ctx, roomID, "Chair 1")
	if err != nil {
		t.Fatalf("create chair: %v", err)
	}

	// create users
	u1, _ := repo.CreateUser(ctx, "chair_user1", "h", "", "")
	u2, _ := repo.CreateUser(ctx, "chair_user2", "h", "", "")

	slot := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Minute)
	expires := time.Now().UTC().Add(15 * time.Minute)

	// user1 holds the chair
	holdID1, _, err := repo.CreateReservationHoldWithChair(ctx, u1, 1, 1, roomID, &chairID, slot, 45, expires)
	if err != nil {
		t.Fatalf("create hold1: %v", err)
	}

	// user2 should not be able to hold same chair overlapping
	if _, _, err := repo.CreateReservationHoldWithChair(ctx, u2, 1, 1, roomID, &chairID, slot.Add(10*time.Minute), 30, expires); err == nil {
		t.Fatalf("expected conflict when holding same chair")
	}

	// confirm hold1
	bookingID, err := repo.ConfirmHold(ctx, u1, holdID1, 0)
	if err != nil {
		t.Fatalf("confirm hold: %v", err)
	}
	if bookingID == 0 {
		t.Fatalf("expected booking id")
	}

	// cleanup best-effort
	pool.Exec(ctx, `DELETE FROM bookings WHERE id=$1`, bookingID)
	pool.Exec(ctx, `DELETE FROM reservation_holds WHERE id=$1`, holdID1)
	pool.Exec(ctx, `DELETE FROM chairs WHERE id=$1`, chairID)
}
