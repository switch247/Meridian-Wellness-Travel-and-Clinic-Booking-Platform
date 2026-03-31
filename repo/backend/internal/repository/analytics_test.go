package repository

import (
	"context"
	"testing"
	"time"

	"meridian/backend/tests/testutil"
)

func TestAnalyticsKPIs_RepurchaseAndRefund(t *testing.T) {
	pool := testutil.DBPoolOrSkip(t)
	repo := New(pool)
	ctx := context.Background()

	// create users
	u1, _ := repo.CreateUser(ctx, "kpi_user1", "h", "", "")
	u2, _ := repo.CreateUser(ctx, "kpi_user2", "h", "", "")

	// insert bookings: u1 two bookings, u2 one booking cancelled
	from := time.Now().UTC().Add(-48 * time.Hour)
	to := time.Now().UTC().Add(48 * time.Hour)
	// booking for u1
	pool.Exec(ctx, `INSERT INTO bookings(hold_id,user_id,package_id,host_id,room_id,slot_start,duration_minutes,status,version) VALUES(0,$1,1,1,1,$2,60,'confirmed',1)`, u1, time.Now().UTC())
	// second booking for u1
	pool.Exec(ctx, `INSERT INTO bookings(hold_id,user_id,package_id,host_id,room_id,slot_start,duration_minutes,status,version) VALUES(0,$1,1,1,1,$2,60,'confirmed',1)`, u1, time.Now().UTC().Add(24*time.Hour))
	// booking for u2 cancelled
	pool.Exec(ctx, `INSERT INTO bookings(hold_id,user_id,package_id,host_id,room_id,slot_start,duration_minutes,status,version) VALUES(0,$1,1,1,1,$2,60,'cancelled',1)`, u2, time.Now().UTC())

	kpis, err := repo.AnalyticsKPIs(ctx, from, to, nil, nil)
	if err != nil {
		t.Fatalf("analytics: %v", err)
	}

	// bookingVolume should be 3
	if bv, ok := kpis["bookingVolume"].(int); !ok || bv < 3 {
		t.Fatalf("expected bookingVolume >=3, got %v", kpis["bookingVolume"])
	}

	// repurchaseRate: only u1 had >1 bookings -> 1/2 buyers = 50%%
	if rr, ok := kpis["repurchaseRate"].(float64); !ok || rr < 49.0 || rr > 51.0 {
		t.Fatalf("unexpected repurchaseRate: %v", kpis["repurchaseRate"])
	}

	// refundRate: 1 cancelled of total bookings
	if rf, ok := kpis["refundRate"].(float64); !ok || rf <= 0.0 {
		t.Fatalf("unexpected refundRate: %v", kpis["refundRate"])
	}
}
