package security_tests

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"meridian/backend/internal/repository"
	"meridian/backend/tests/testutil"
)

func TestBookingStatusUpdateRejectsCrossTenantLocation(t *testing.T) {
	pool := testutil.DBPoolOrSkip(t)
	defer pool.Close()
	ctx := context.Background()
	repo := repository.New(pool)

	var defaultLocationID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM locations WHERE name='Default Location' LIMIT 1`).Scan(&defaultLocationID); err != nil {
		t.Fatalf("default location: %v", err)
	}

	locationName := fmt.Sprintf("tenant-isolation-%s", t.Name())
	var altLocationID int64
	if err := pool.QueryRow(ctx, `INSERT INTO locations(name) VALUES($1) RETURNING id`, locationName).Scan(&altLocationID); err != nil {
		t.Fatalf("alt location: %v", err)
	}

	var roomID int64
	roomName := fmt.Sprintf("tenant-room-%s", t.Name())
	if err := pool.QueryRow(ctx, `INSERT INTO rooms(name,chairs_count,location_id) VALUES($1,1,$2) RETURNING id`, roomName, altLocationID).Scan(&roomID); err != nil {
		t.Fatalf("insert room: %v", err)
	}

	var travelerID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM users WHERE username='traveler1@example.com' LIMIT 1`).Scan(&travelerID); err != nil {
		t.Fatalf("traveler id: %v", err)
	}
	var hostID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM users WHERE username='coach@example.com' LIMIT 1`).Scan(&hostID); err != nil {
		t.Fatalf("host id: %v", err)
	}

	_, _ = pool.Exec(ctx, `INSERT INTO user_locations(user_id, location_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, hostID, altLocationID)
	_, _ = pool.Exec(ctx, `INSERT INTO user_locations(user_id, location_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, travelerID, altLocationID)

	var packageID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM packages LIMIT 1`).Scan(&packageID); err != nil {
		t.Fatalf("package id: %v", err)
	}

	slot := time.Now().UTC().Add(2 * time.Hour)
	expiry := slot.Add(30 * time.Minute)
	var holdID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO reservation_holds(user_id,package_id,host_id,room_id,slot_start,duration_minutes,status,version,expires_at)
		VALUES($1,$2,$3,$4,$5,$6,'scheduled',1,$7)
		RETURNING id
	`, travelerID, packageID, hostID, roomID, slot, 60, expiry).Scan(&holdID); err != nil {
		t.Fatalf("insert hold: %v", err)
	}

	var bookingID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO bookings(hold_id,user_id,package_id,host_id,room_id,slot_start,duration_minutes,status,version)
		VALUES($1,$2,$3,$4,$5,$6,$7,'confirmed',1)
		RETURNING id
	`, holdID, travelerID, packageID, hostID, roomID, slot, 60).Scan(&bookingID); err != nil {
		t.Fatalf("insert booking: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM bookings WHERE id=$1`, bookingID)
		_, _ = pool.Exec(ctx, `DELETE FROM reservation_holds WHERE id=$1`, holdID)
		_, _ = pool.Exec(ctx, `DELETE FROM rooms WHERE id=$1`, roomID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_locations WHERE location_id=$1`, altLocationID)
		_, _ = pool.Exec(ctx, `DELETE FROM locations WHERE id=$1`, altLocationID)
	})

	_, _, err := repo.UpdateBookingStatusByLocation(ctx, bookingID, defaultLocationID, "cancelled", nil)
	if err == nil {
		t.Fatalf("expected location guard failure")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
