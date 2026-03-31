package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateChair creates a chair record for a room and returns its id.
func (r *Repository) CreateChair(ctx context.Context, roomID int64, name string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
        INSERT INTO chairs(room_id,name) VALUES($1,$2) RETURNING id
    `, roomID, name).Scan(&id)
	return id, err
}

// GetChairsByRoom returns chairs for a room.
func (r *Repository) GetChairsByRoom(ctx context.Context, roomID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,created_at FROM chairs WHERE room_id=$1 ORDER BY id`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var name string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "name": name, "createdAt": createdAt})
	}
	return out, rows.Err()
}

// GetAvailableChairsForSlot returns chair ids that are free for the slot (no active holds or confirmed bookings).
func (r *Repository) GetAvailableChairsForSlot(ctx context.Context, hostID, roomID int64, slotStart time.Time, duration int) ([]int64, error) {
	slotEnd := slotStart.Add(time.Duration(duration) * time.Minute)
	rows, err := r.pool.Query(ctx, `
        SELECT c.id FROM chairs c
        WHERE c.room_id=$1
        AND NOT EXISTS(
            SELECT 1 FROM reservation_holds h
            WHERE h.chair_id=c.id AND h.status='active' AND h.expires_at > NOW()
            AND h.slot_start < $3 AND (h.slot_start + (h.duration_minutes || ' minutes')::interval) > $2
        )
        AND NOT EXISTS(
            SELECT 1 FROM bookings b
            WHERE b.chair_id=c.id AND b.status='confirmed'
            AND b.slot_start < $3 AND (b.slot_start + (b.duration_minutes || ' minutes')::interval) > $2
        )
    `, roomID, slotStart, slotEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CreateReservationHoldWithChair behaves like CreateReservationHold but reserves a specific chair.
func (r *Repository) CreateReservationHoldWithChair(ctx context.Context, userID, packageID, hostID, roomID int64, chairID *int64, slotStart time.Time, durationMinutes int, expiresAt time.Time) (int64, int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	var quota int
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, hostID+1_000_000_000); err != nil {
		return 0, 0, err
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, roomID+2_000_000_000); err != nil {
		return 0, 0, err
	}
	if chairID != nil {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, *chairID+3_000_000_000); err != nil {
			return 0, 0, err
		}
	}
	if err := tx.QueryRow(ctx, `SELECT inventory_remaining FROM package_calendar WHERE package_id=$1 AND service_date=$2 FOR UPDATE`, packageID, slotStart.Format("2006-01-02")).Scan(&quota); err != nil {
		return 0, 0, err
	}
	if quota <= 0 {
		return 0, 0, fmt.Errorf("no inventory")
	}

	slotEnd := slotStart.Add(time.Duration(durationMinutes) * time.Minute)

	// Check holds/bookings conflicts for chair if provided, otherwise fallback to host/room checks.
	var conflict bool
	if chairID != nil {
		err = tx.QueryRow(ctx, `
            SELECT EXISTS(
                SELECT 1 FROM reservation_holds
                WHERE status='active' AND expires_at > NOW()
                AND slot_start < $3
                AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
				AND (host_id=$2 OR chair_id=$4)
            )
        `, slotStart, hostID, slotEnd, *chairID).Scan(&conflict)
		if err != nil {
			return 0, 0, err
		}
		if conflict {
			return 0, 0, fmt.Errorf("slot unavailable")
		}
		if err := tx.QueryRow(ctx, `
            SELECT EXISTS(
                SELECT 1 FROM bookings
                WHERE status='confirmed'
                AND slot_start < $3
                AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
				AND (host_id=$2 OR chair_id=$4)
            )
        `, slotStart, hostID, slotEnd, *chairID).Scan(&conflict); err != nil {
			return 0, 0, err
		}
		if conflict {
			return 0, 0, fmt.Errorf("slot unavailable")
		}
	} else {
		err = tx.QueryRow(ctx, `
            SELECT EXISTS(
                SELECT 1 FROM reservation_holds
                WHERE status='active' AND expires_at > NOW()
                AND slot_start < $4
                AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
                AND (host_id=$2 OR room_id=$3)
            )
        `, slotStart, hostID, roomID, slotEnd).Scan(&conflict)
		if err != nil {
			return 0, 0, err
		}
		if conflict {
			return 0, 0, fmt.Errorf("slot unavailable")
		}
		if err := tx.QueryRow(ctx, `
            SELECT EXISTS(
                SELECT 1 FROM bookings
                WHERE status='confirmed'
                AND slot_start < $4
                AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
                AND (host_id=$2 OR room_id=$3)
            )
        `, slotStart, hostID, roomID, slotEnd).Scan(&conflict); err != nil {
			return 0, 0, err
		}
		if conflict {
			return 0, 0, fmt.Errorf("slot unavailable")
		}
	}

	var holdID int64
	var version int
	if chairID != nil {
		err = tx.QueryRow(ctx, `
            INSERT INTO reservation_holds(user_id,package_id,host_id,room_id,chair_id,slot_start,duration_minutes,expires_at,status,version)
            VALUES($1,$2,$3,$4,$5,$6,$7,$8,'active',1)
            RETURNING id,version
        `, userID, packageID, hostID, roomID, *chairID, slotStart, durationMinutes, expiresAt).Scan(&holdID, &version)
	} else {
		err = tx.QueryRow(ctx, `
            INSERT INTO reservation_holds(user_id,package_id,host_id,room_id,slot_start,duration_minutes,expires_at,status,version)
            VALUES($1,$2,$3,$4,$5,$6,$7,'active',1)
            RETURNING id,version
        `, userID, packageID, hostID, roomID, slotStart, durationMinutes, expiresAt).Scan(&holdID, &version)
	}
	if err != nil {
		if errorsIsNoRows(err) {
			// fallthrough
		}
		return 0, 0, err
	}

	if _, err := tx.Exec(ctx, `UPDATE package_calendar SET inventory_remaining=inventory_remaining-1, version=version+1 WHERE package_id=$1 AND service_date=$2 AND inventory_remaining>0`, packageID, slotStart.Format("2006-01-02")); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return holdID, version, nil
}

// small compatibility helper to detect pgx ErrNoRows across drivers
func errorsIsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
