package repository

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"meridian/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound               = errors.New("not found")
	ErrHoldExpired            = errors.New("hold expired")
	ErrAddressRequired        = errors.New("address required")
	ErrPostalBlocked          = errors.New("address not serviceable")
	ErrServiceWindowViolation = errors.New("service not available at requested time window")
	ErrInvalidBookingStatus   = errors.New("invalid booking status")
	validBookingStatuses      = map[string]struct{}{
		"scheduled":   {},
		"confirmed":   {},
		"checked_in":  {},
		"in_progress": {},
		"completed":   {},
		"cancelled":   {},
	}
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func isBookingStatusValid(status string) bool {
	_, ok := validBookingStatuses[status]
	return ok
}

func (r *Repository) CreateUser(ctx context.Context, username, passwordHash, encryptedPhone, encryptedAddress string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users(username,password_hash,encrypted_phone,encrypted_address)
		VALUES($1,$2,$3,$4)
		RETURNING id
	`, username, passwordHash, encryptedPhone, encryptedAddress).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) FindUserByUsername(ctx context.Context, username string) (domain.User, []string, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT id,username,password_hash,failed_attempts,locked_until,encrypted_phone,encrypted_address,created_at,last_password_reset
		FROM users WHERE username=$1
	`, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.FailedAttempts, &u.LockedUntil, &u.EncryptedPhone, &u.EncryptedAddress, &u.CreatedAt, &u.LastPasswordReset,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, nil, ErrNotFound
		}
		return domain.User{}, nil, err
	}
	roles, err := r.GetRolesForUser(ctx, u.ID)
	if err != nil {
		return domain.User{}, nil, err
	}
	return u, roles, nil
}

func (r *Repository) GetRolesForUser(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT role_name FROM user_roles WHERE user_id=$1 ORDER BY role_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) SetUserFailedAttempts(ctx context.Context, userID int64, attempts int, lockedUntil *time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET failed_attempts=$1, locked_until=$2 WHERE id=$3`, attempts, lockedUntil, userID)
	return err
}

func (r *Repository) ResetFailedAttempts(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET failed_attempts=0, locked_until=NULL WHERE id=$1`, userID)
	return err
}

func (r *Repository) AssignRole(ctx context.Context, actorID, targetID int64, role string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	before, err := rolesAsCSV(ctx, tx, targetID)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO user_roles(user_id,role_name) VALUES($1,$2) ON CONFLICT DO NOTHING`, targetID, role); err != nil {
		return err
	}

	after, err := rolesAsCSV(ctx, tx, targetID)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO permission_audits(actor_id,target_user_id,action,before_state,after_state)
		VALUES($1,$2,'assign_role',$3,$4)
	`, actorID, targetID, before, after); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func rolesAsCSV(ctx context.Context, tx pgx.Tx, userID int64) (string, error) {
	rows, err := tx.Query(ctx, `SELECT role_name FROM user_roles WHERE user_id=$1 ORDER BY role_name`, userID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	roles := ""
	first := true
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return "", err
		}
		if !first {
			roles += ","
		}
		roles += role
		first = false
	}
	return roles, rows.Err()
}

func (r *Repository) CreateAddress(ctx context.Context, userID int64, line1, line2, city, state, postal, normalized string, inCoverage, duplicate bool, line1Encrypted, line2Encrypted string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO addresses(user_id,line1,line2,city,state,postal_code,normalized_key,in_coverage,is_duplicate,line1_encrypted,line2_encrypted)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, userID, line1, line2, city, state, postal, normalized, inCoverage, duplicate, line1Encrypted, line2Encrypted)
	return err
}

func (r *Repository) AddressExistsByNormalizedKey(ctx context.Context, userID int64, normalized string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM addresses WHERE user_id=$1 AND normalized_key=$2)`, userID, normalized).Scan(&exists)
	return exists, err
}

func (r *Repository) ListAddressesByUser(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,line1,line2,city,state,postal_code,normalized_key,in_coverage,is_duplicate,created_at,line1_encrypted,line2_encrypted
		FROM addresses WHERE user_id=$1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var line1, line2, city, state, postal, normalized string
		var line1Encrypted, line2Encrypted string
		var inCoverage, duplicate bool
		var createdAt time.Time
		if err := rows.Scan(&id, &line1, &line2, &city, &state, &postal, &normalized, &inCoverage, &duplicate, &createdAt, &line1Encrypted, &line2Encrypted); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":             id,
			"line1":          line1,
			"line2":          line2,
			"city":           city,
			"state":          state,
			"postalCode":     postal,
			"normalizedKey":  normalized,
			"inCoverage":     inCoverage,
			"duplicate":      duplicate,
			"createdAt":      createdAt,
			"line1Encrypted": line1Encrypted,
			"line2Encrypted": line2Encrypted,
		})
	}
	return items, rows.Err()
}

func (r *Repository) CreateContact(ctx context.Context, userID int64, name, relationship, phoneMasked, phoneEncrypted string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO profile_contacts(user_id,name,relationship,phone_masked,phone_encrypted)
		VALUES($1,$2,$3,$4,$5)
		RETURNING id
	`, userID, name, relationship, phoneMasked, phoneEncrypted).Scan(&id)
	return id, err
}

func (r *Repository) ListContactsByUser(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,name,relationship,phone_masked,created_at
		FROM profile_contacts WHERE user_id=$1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var name, relationship, phoneMasked string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &relationship, &phoneMasked, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":           id,
			"name":         name,
			"relationship": relationship,
			"phoneMasked":  phoneMasked,
			"createdAt":    createdAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) GetPrimaryPostalCodeForUser(ctx context.Context, userID int64) (string, error) {
	var postal string
	err := r.pool.QueryRow(ctx, `SELECT postal_code FROM addresses WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1`, userID).Scan(&postal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrAddressRequired
		}
		return "", err
	}
	return postal, nil
}

func (r *Repository) EnforceServiceRules(ctx context.Context, userID int64, slotStart time.Time) error {
	postal, err := r.GetPrimaryPostalCodeForUser(ctx, userID)
	if err != nil {
		return err
	}

	var blocked bool
	var start sql.NullTime
	var end sql.NullTime
	err = r.pool.QueryRow(ctx, `
		SELECT sr.blocked, sr.start_time, sr.end_time
		FROM blocked_postal_codes b
		JOIN service_rules sr ON sr.id = b.service_rule_id
		WHERE b.postal_code=$1
		ORDER BY sr.updated_at DESC
		LIMIT 1
	`, postal).Scan(&blocked, &start, &end)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if blocked {
		return ErrPostalBlocked
	}
	if !start.Valid || !end.Valid {
		return nil
	}
	if !timeWindowAllows(slotStart, start.Time, end.Time) {
		return ErrServiceWindowViolation
	}
	return nil
}

func timeWindowAllows(slot, start, end time.Time) bool {
	slotUTC := slot.UTC()
	startUTC := start.UTC()
	endUTC := end.UTC()
	slotTOD := time.Date(0, time.January, 1, slotUTC.Hour(), slotUTC.Minute(), slotUTC.Second(), slotUTC.Nanosecond(), time.UTC)
	startTOD := time.Date(0, time.January, 1, startUTC.Hour(), startUTC.Minute(), startUTC.Second(), startUTC.Nanosecond(), time.UTC)
	endTOD := time.Date(0, time.January, 1, endUTC.Hour(), endUTC.Minute(), endUTC.Second(), endUTC.Nanosecond(), time.UTC)
	if startTOD.Before(endTOD) || startTOD.Equal(endTOD) {
		return (slotTOD.Equal(startTOD) || slotTOD.After(startTOD)) && slotTOD.Before(endTOD)
	}
	return slotTOD.Equal(startTOD) || slotTOD.After(startTOD) || slotTOD.Before(endTOD)
}

func (r *Repository) ListRegions(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,name,parent_region_id,description,active,created_at
		FROM regions ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var name, description string
		var active bool
		var createdAt time.Time
		var parent sql.NullInt64
		if err := rows.Scan(&id, &name, &parent, &description, &active, &createdAt); err != nil {
			return nil, err
		}
		var parentID *int64
		if parent.Valid {
			temp := parent.Int64
			parentID = &temp
		}
		out = append(out, map[string]any{
			"id":          id,
			"name":        name,
			"parentId":    parentID,
			"description": description,
			"active":      active,
			"createdAt":   createdAt,
		})
	}
	return out, rows.Err()
}

func (r *Repository) CreateRegion(ctx context.Context, name, description string, parentID *int64) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO regions(name,description,parent_region_id)
		VALUES($1,$2,$3)
		RETURNING id
	`, name, description, parentID).Scan(&id)
	return id, err
}

func (r *Repository) UpsertServiceRule(ctx context.Context, regionID int64, allowPickup, allowMail, blocked bool, start, end *time.Time) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO service_rules(region_id,allow_home_pickup,allow_mail_documents,blocked,start_time,end_time,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,NOW(),NOW())
		ON CONFLICT(region_id) DO UPDATE
		SET allow_home_pickup=EXCLUDED.allow_home_pickup,
			allow_mail_documents=EXCLUDED.allow_mail_documents,
			blocked=EXCLUDED.blocked,
			start_time=EXCLUDED.start_time,
			end_time=EXCLUDED.end_time,
			updated_at=NOW()
		RETURNING id
	`, regionID, allowPickup, allowMail, blocked, start, end).Scan(&id)
	return id, err
}

func (r *Repository) ListBlockedPostalCodes(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.id,b.postal_code,sr.id,sr.region_id,r.name
		FROM blocked_postal_codes b
		JOIN service_rules sr ON sr.id = b.service_rule_id
		JOIN regions r ON r.id = sr.region_id
		ORDER BY b.postal_code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, ruleID, regionID int64
		var postal, regionName string
		if err := rows.Scan(&id, &postal, &ruleID, &regionID, &regionName); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":            id,
			"postalCode":    postal,
			"serviceRuleId": ruleID,
			"regionId":      regionID,
			"regionName":    regionName,
		})
	}
	return items, rows.Err()
}

func (r *Repository) AddBlockedPostalCode(ctx context.Context, serviceRuleID int64, postalCode string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO blocked_postal_codes(service_rule_id,postal_code,created_at)
		VALUES($1,$2,NOW())
		ON CONFLICT(service_rule_id,postal_code) DO NOTHING
		RETURNING id
	`, serviceRuleID, postalCode).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = r.pool.QueryRow(ctx, `
				SELECT id FROM blocked_postal_codes WHERE service_rule_id=$1 AND postal_code=$2
			`, serviceRuleID, postalCode).Scan(&id)
		} else {
			return 0, err
		}
	}
	return id, err
}

func (r *Repository) DeleteContact(ctx context.Context, userID, contactID int64) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM profile_contacts WHERE id=$1 AND user_id=$2`, contactID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID int64) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT id,username,password_hash,failed_attempts,locked_until,encrypted_phone,encrypted_address,created_at,last_password_reset
		FROM users WHERE id=$1
	`, userID).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.FailedAttempts, &u.LockedUntil, &u.EncryptedPhone, &u.EncryptedAddress, &u.CreatedAt, &u.LastPasswordReset,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *Repository) CreateReservationHold(ctx context.Context, userID, packageID, hostID, roomID int64, slotStart time.Time, durationMinutes int, expiresAt time.Time) (int64, int, error) {
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
	if err := tx.QueryRow(ctx, `SELECT inventory_remaining FROM package_calendar WHERE package_id=$1 AND service_date=$2 FOR UPDATE`, packageID, slotStart.Format("2006-01-02")).Scan(&quota); err != nil {
		return 0, 0, err
	}
	if quota <= 0 {
		return 0, 0, fmt.Errorf("no inventory")
	}

	var conflict bool
	slotEnd := slotStart.Add(time.Duration(durationMinutes) * time.Minute)
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM reservation_holds
			WHERE status='active'
			AND expires_at > NOW()
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

	var holdID int64
	var version int
	err = tx.QueryRow(ctx, `
		INSERT INTO reservation_holds(user_id,package_id,host_id,room_id,slot_start,duration_minutes,expires_at,status,version)
		VALUES($1,$2,$3,$4,$5,$6,$7,'active',1)
		RETURNING id,version
	`, userID, packageID, hostID, roomID, slotStart, durationMinutes, expiresAt).Scan(&holdID, &version)
	if err != nil {
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

func (r *Repository) ReleaseExpiredHolds(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		WITH expired AS (
			UPDATE reservation_holds
			SET status='released'
			WHERE status='active' AND expires_at <= NOW()
			RETURNING package_id, slot_start
		)
		UPDATE package_calendar pc
		SET inventory_remaining=inventory_remaining+1, version=version+1
		FROM expired e
		WHERE pc.package_id=e.package_id AND pc.service_date=e.slot_start::date
	`)
	return err
}

func (r *Repository) ListHoldsByUser(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,package_id,host_id,room_id,slot_start,duration_minutes,expires_at,status,version,created_at
		FROM reservation_holds WHERE user_id=$1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, packageID, hostID, roomID int64
		var slotStart, expiresAt, createdAt time.Time
		var duration, version int
		var status string
		if err := rows.Scan(&id, &packageID, &hostID, &roomID, &slotStart, &duration, &expiresAt, &status, &version, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":              id,
			"packageId":       packageID,
			"hostId":          hostID,
			"roomId":          roomID,
			"slotStart":       slotStart,
			"durationMinutes": duration,
			"expiresAt":       expiresAt,
			"status":          status,
			"version":         version,
			"createdAt":       createdAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) DeleteAddress(ctx context.Context, userID, addressID int64) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM addresses WHERE id=$1 AND user_id=$2`, addressID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CancelHoldByUser(ctx context.Context, userID, holdID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var packageID int64
	var slotStart time.Time
	var status string
	err = tx.QueryRow(ctx, `SELECT package_id, slot_start, status, user_id FROM reservation_holds WHERE id=$1`, holdID).Scan(&packageID, &slotStart, &status, new(interface{}))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if status != "active" {
		return fmt.Errorf("hold not active")
	}

	// Ensure owner is cancelling: we require caller to pass userID and verify match
	var ownerID int64
	if err := tx.QueryRow(ctx, `SELECT user_id FROM reservation_holds WHERE id=$1`, holdID).Scan(&ownerID); err != nil {
		return err
	}
	if ownerID != userID {
		return fmt.Errorf("not owner")
	}

	if _, err := tx.Exec(ctx, `UPDATE reservation_holds SET status='cancelled' WHERE id=$1`, holdID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE package_calendar SET inventory_remaining=inventory_remaining+1, version=version+1 WHERE package_id=$1 AND service_date=$2`, packageID, slotStart.Format("2006-01-02")); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListBookingHistoryByUser(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,package_id,host_id,room_id,slot_start,duration_minutes,status,created_at
		FROM reservation_holds WHERE user_id=$1 AND status<>'active' ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, packageID, hostID, roomID int64
		var slotStart, createdAt time.Time
		var duration int
		var status string
		if err := rows.Scan(&id, &packageID, &hostID, &roomID, &slotStart, &duration, &status, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":              id,
			"packageId":       packageID,
			"hostId":          hostID,
			"roomId":          roomID,
			"slotStart":       slotStart,
			"durationMinutes": duration,
			"status":          status,
			"createdAt":       createdAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) GetBookingHost(ctx context.Context, bookingID int64) (int64, error) {
	var hostID int64
	err := r.pool.QueryRow(ctx, `
		SELECT h.host_id
		FROM reservation_holds h
		JOIN bookings b ON b.hold_id=h.id
		WHERE b.id=$1
	`, bookingID).Scan(&hostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return hostID, nil
}

func (r *Repository) ListUsers(ctx context.Context, roleFilter string) ([]map[string]any, error) {
	query := `SELECT id, username, created_at FROM users`
	args := []any{}
	if roleFilter != "" {
		query += ` WHERE EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id=users.id AND ur.role_name=$1)`
		args = append(args, roleFilter)
	}
	query += ` ORDER BY id DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var username string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &createdAt); err != nil {
			return nil, err
		}
		roles, err := r.GetRolesForUser(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":        id,
			"username":  username,
			"roles":     roles,
			"createdAt": createdAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) ListPermissionAudits(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,actor_id,target_user_id,action,before_state,after_state,created_at
		FROM permission_audits ORDER BY created_at DESC LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, actorID, targetID int64
		var action, before, after string
		var createdAt time.Time
		if err := rows.Scan(&id, &actorID, &targetID, &action, &before, &after, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":           id,
			"actorId":      actorID,
			"targetUserId": targetID,
			"action":       action,
			"before":       before,
			"after":        after,
			"createdAt":    createdAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) ListHostAgenda(ctx context.Context, hostID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.id, h.user_id, h.package_id, h.room_id, h.slot_start, h.duration_minutes, h.status,
		       b.id, b.session_notes_encrypted
		FROM reservation_holds h
		LEFT JOIN bookings b ON b.hold_id=h.id
		WHERE h.host_id=$1 AND h.status<>'active'
		ORDER BY h.slot_start DESC
		LIMIT 200
	`, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, userID, packageID, roomID, bookingID sql.NullInt64
		var slotStart time.Time
		var duration int
		var status string
		var rawNotes sql.NullString
		if err := rows.Scan(&id, &userID, &packageID, &roomID, &slotStart, &duration, &status, &bookingID, &rawNotes); err != nil {
			return nil, err
		}
		item := map[string]any{
			"id":              id.Int64,
			"travelerId":      userID.Int64,
			"packageId":       packageID.Int64,
			"roomId":          roomID.Int64,
			"slotStart":       slotStart,
			"durationMinutes": duration,
			"status":          status,
		}
		if bookingID.Valid {
			item["bookingId"] = bookingID.Int64
		}
		if rawNotes.Valid {
			item["sessionNotesEncrypted"] = rawNotes.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListRoomAgenda(ctx context.Context, roomID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.id, h.user_id, h.package_id, h.host_id, h.slot_start, h.duration_minutes, h.status,
		       b.id, b.session_notes_encrypted
		FROM reservation_holds h
		LEFT JOIN bookings b ON b.hold_id=h.id
		WHERE h.room_id=$1 AND h.status<>'active'
		ORDER BY h.slot_start DESC
		LIMIT 200
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, userID, packageID, hostID, bookingID sql.NullInt64
		var slotStart time.Time
		var duration int
		var status string
		var rawNotes sql.NullString
		if err := rows.Scan(&id, &userID, &packageID, &hostID, &slotStart, &duration, &status, &bookingID, &rawNotes); err != nil {
			return nil, err
		}
		item := map[string]any{
			"id":              id.Int64,
			"travelerId":      userID.Int64,
			"packageId":       packageID.Int64,
			"hostId":          hostID.Int64,
			"slotStart":       slotStart,
			"durationMinutes": duration,
			"status":          status,
		}
		if bookingID.Valid {
			item["bookingId"] = bookingID.Int64
		}
		if rawNotes.Valid {
			item["sessionNotesEncrypted"] = rawNotes.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListCatalog(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, p.published, d.name destination, c.service_date, c.price_cents, c.inventory_remaining, c.blackout_note
		FROM packages p
		JOIN destinations d ON d.id=p.destination_id
		LEFT JOIN package_calendar c ON c.package_id=p.id
		WHERE p.published=true
		ORDER BY p.id, c.service_date
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var name, destination string
		var published bool
		var serviceDate *time.Time
		var priceCents *int
		var inventory *int
		var blackout *string
		if err := rows.Scan(&id, &name, &published, &destination, &serviceDate, &priceCents, &inventory, &blackout); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":                 id,
			"name":               name,
			"published":          published,
			"destination":        destination,
			"serviceDate":        serviceDate,
			"priceCents":         priceCents,
			"inventoryRemaining": inventory,
			"blackoutNote":       blackout,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range items {
		if items[i]["blackoutNote"] == nil {
			items[i]["blackoutNote"] = ""
		}
		if s, ok := items[i]["destination"].(string); ok {
			items[i]["destination"] = strings.TrimSpace(s)
		}
	}
	return items, nil
}

func (r *Repository) ListRoutes(ctx context.Context) ([]map[string]any, error) {
	return r.listRichCatalogByTable(ctx, "routes")
}

func (r *Repository) ListHotels(ctx context.Context) ([]map[string]any, error) {
	return r.listRichCatalogByTable(ctx, "hotels")
}

func (r *Repository) ListAttractions(ctx context.Context) ([]map[string]any, error) {
	return r.listRichCatalogByTable(ctx, "attractions")
}

func (r *Repository) listRichCatalogByTable(ctx context.Context, table string) ([]map[string]any, error) {
	q := fmt.Sprintf(`
		SELECT id, destination_id, name, rich_description, image_paths, published, created_at
		FROM %s
		WHERE published=true
		ORDER BY id DESC
	`, table)
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, destinationID int64
		var name, richDescription string
		var imagePaths []string
		var published bool
		var createdAt time.Time
		if err := rows.Scan(&id, &destinationID, &name, &richDescription, &imagePaths, &published, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":              id,
			"destinationId":   destinationID,
			"name":            name,
			"richDescription": richDescription,
			"imagePaths":      imagePaths,
			"published":       published,
			"createdAt":       createdAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) ListAvailableSlots(ctx context.Context, hostID, roomID int64, day time.Time, duration int, granularity int) ([]map[string]any, error) {
	if err := r.ReleaseExpiredHolds(ctx); err != nil {
		return nil, err
	}
	weekday := int(day.Weekday())
	var holidayClosed bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM holidays WHERE holiday_date=$1 AND closed_all_day=true)`, day.Format("2006-01-02")).Scan(&holidayClosed); err != nil {
		return nil, err
	}
	if holidayClosed {
		return []map[string]any{}, nil
	}

	var startTime, endTime time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT start_time, end_time
		FROM host_availability
		WHERE host_id=$1 AND active=true AND weekday=$2 AND (room_id IS NULL OR room_id=$3)
		ORDER BY id DESC
		LIMIT 1
	`, hostID, weekday, roomID).Scan(&startTime, &endTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	var hasException bool
	var exceptionAvailable bool
	var exceptionStart, exceptionEnd *time.Time
	if err := r.pool.QueryRow(ctx, `
		SELECT true, is_available, start_time, end_time
		FROM host_availability_exceptions
		WHERE host_id=$1 AND exception_date=$2
		ORDER BY id DESC
		LIMIT 1
	`, hostID, day.Format("2006-01-02")).Scan(&hasException, &exceptionAvailable, &exceptionStart, &exceptionEnd); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	base := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	start := time.Date(base.Year(), base.Month(), base.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	end := time.Date(base.Year(), base.Month(), base.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)
	if hasException {
		if !exceptionAvailable {
			return []map[string]any{}, nil
		}
		if exceptionStart != nil && exceptionEnd != nil {
			start = time.Date(base.Year(), base.Month(), base.Day(), exceptionStart.Hour(), exceptionStart.Minute(), 0, 0, time.UTC)
			end = time.Date(base.Year(), base.Month(), base.Day(), exceptionEnd.Hour(), exceptionEnd.Minute(), 0, 0, time.UTC)
		}
	}

	slots := []map[string]any{}
	step := time.Duration(granularity) * time.Minute
	block := time.Duration(duration) * time.Minute
	for t := start; t.Add(block).Before(end) || t.Add(block).Equal(end); t = t.Add(step) {
		var conflict bool
		candidateEnd := t.Add(block)
		if err := r.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM reservation_holds
				WHERE status='active'
				AND expires_at > NOW()
				AND slot_start < $4
				AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
				AND (host_id=$2 OR room_id=$3)
			)
		`, t, hostID, roomID, candidateEnd).Scan(&conflict); err != nil {
			return nil, err
		}
		if !conflict {
			if err := r.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM bookings
					WHERE status='confirmed'
					AND slot_start < $4
					AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
					AND (host_id=$2 OR room_id=$3)
				)
			`, t, hostID, roomID, candidateEnd).Scan(&conflict); err != nil {
				return nil, err
			}
		}
		if !conflict {
			slots = append(slots, map[string]any{
				"slotStart":       t,
				"slotEnd":         candidateEnd,
				"durationMinutes": duration,
			})
		}
	}
	return slots, nil
}

func (r *Repository) ConfirmHold(ctx context.Context, userID, holdID int64, expectedVersion int) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var packageID, hostID, roomID int64
	var slotStart time.Time
	var duration, version int
	var status string
	var expiresAt time.Time
	var ownerID int64
	err = tx.QueryRow(ctx, `
		SELECT package_id, host_id, room_id, slot_start, duration_minutes, version, status, expires_at, user_id
		FROM reservation_holds WHERE id=$1 FOR UPDATE
	`, holdID).Scan(&packageID, &hostID, &roomID, &slotStart, &duration, &version, &status, &expiresAt, &ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	if ownerID != userID {
		return 0, fmt.Errorf("not owner")
	}
	if status != "active" {
		return 0, fmt.Errorf("hold not active")
	}
	if expectedVersion > 0 && version != expectedVersion {
		return 0, fmt.Errorf("version conflict")
	}
	if expiresAt.Before(time.Now().UTC()) {
		return 0, ErrHoldExpired
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, hostID+1_000_000_000); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, roomID+2_000_000_000); err != nil {
		return 0, err
	}
	slotEnd := slotStart.Add(time.Duration(duration) * time.Minute)
	var conflict bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE status='confirmed'
			AND id <> COALESCE((SELECT id FROM bookings WHERE hold_id=$5), -1)
			AND slot_start < $4
			AND (slot_start + (duration_minutes || ' minutes')::interval) > $1
			AND (host_id=$2 OR room_id=$3)
		)
	`, slotStart, hostID, roomID, slotEnd, holdID).Scan(&conflict); err != nil {
		return 0, err
	}
	if conflict {
		return 0, fmt.Errorf("slot unavailable")
	}

	var bookingID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO bookings(hold_id,user_id,package_id,host_id,room_id,slot_start,duration_minutes,status,version)
		VALUES($1,$2,$3,$4,$5,$6,$7,'scheduled',1)
		RETURNING id
	`, holdID, userID, packageID, hostID, roomID, slotStart, duration).Scan(&bookingID); err != nil {
		return 0, err
	}

	if _, err := tx.Exec(ctx, `UPDATE reservation_holds SET status='scheduled', version=version+1 WHERE id=$1`, holdID); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return bookingID, nil
}

func (r *Repository) UpdateBookingStatus(ctx context.Context, bookingID int64, status string, encryptedNotes *string) error {
	if !isBookingStatusValid(status) {
		return ErrInvalidBookingStatus
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var holdID int64
	if err := tx.QueryRow(ctx, `SELECT hold_id FROM bookings WHERE id=$1 FOR UPDATE`, bookingID).Scan(&holdID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	args := []any{status}
	query := `UPDATE bookings SET status=$1`
	if encryptedNotes != nil {
		args = append(args, *encryptedNotes)
		query += `, session_notes_encrypted=$2`
	}
	args = append(args, bookingID)
	query += fmt.Sprintf(` WHERE id=$%d`, len(args))

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reservation_holds SET status=$1, version=version+1 WHERE id=$2`, status, holdID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListCommunityPosts(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, author_user_id, title, body, destination_id, provider_user_id, status, created_at FROM community_posts WHERE status='active' ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, authorID int64
		var title, body, status string
		var destinationID, providerID *int64
		var createdAt time.Time
		if err := rows.Scan(&id, &authorID, &title, &body, &destinationID, &providerID, &status, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "authorUserId": authorID, "title": title, "body": body, "destinationId": destinationID, "providerUserId": providerID, "status": status, "createdAt": createdAt})
	}
	return items, rows.Err()
}

func (r *Repository) CreateCommunityPost(ctx context.Context, authorID int64, title, body string, destinationID, providerID *int64) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO community_posts(author_user_id,title,body,destination_id,provider_user_id,status)
		VALUES($1,$2,$3,$4,$5,'active')
		RETURNING id
	`, authorID, title, body, destinationID, providerID).Scan(&id)
	return id, err
}

func (r *Repository) ListCommentsByPost(ctx context.Context, postID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, post_id, author_user_id, parent_comment_id, body, status, created_at
		FROM community_comments
		WHERE post_id=$1 AND status='active'
		ORDER BY created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, pID, authorID int64
		var parentID *int64
		var body, status string
		var createdAt time.Time
		if err := rows.Scan(&id, &pID, &authorID, &parentID, &body, &status, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "postId": pID, "authorUserId": authorID, "parentCommentId": parentID, "body": body, "status": status, "createdAt": createdAt})
	}
	return items, rows.Err()
}

func (r *Repository) CreateComment(ctx context.Context, authorID, postID int64, parentID *int64, body string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO community_comments(post_id,author_user_id,parent_comment_id,body,status)
		VALUES($1,$2,$3,$4,'active')
		RETURNING id
	`, postID, authorID, parentID, body).Scan(&id)
	return id, err
}

func (r *Repository) ToggleFavorite(ctx context.Context, userID, packageID int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO community_favorites(user_id, package_id) VALUES($1,$2)
		ON CONFLICT (user_id, package_id) DO NOTHING
	`, userID, packageID)
	return err
}

func (r *Repository) FollowUser(ctx context.Context, followerID, targetID int64) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO user_follows(follower_user_id,target_user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, followerID, targetID)
	return err
}

func (r *Repository) BlockUser(ctx context.Context, blockerID, blockedID int64) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO user_blocks(blocker_user_id,blocked_user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, blockerID, blockedID)
	return err
}

func (r *Repository) ReportTarget(ctx context.Context, reporterID int64, targetType string, targetID int64, reason string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO moderation_reports(reporter_user_id,target_type,target_id,reason,status)
		VALUES($1,$2,$3,$4,'pending')
		RETURNING id
	`, reporterID, targetType, targetID, reason).Scan(&id)
	return id, err
}

func (r *Repository) ResolveReport(ctx context.Context, reportID, resolverID int64, status, outcome string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE moderation_reports
		SET status=$2, outcome_note=$3, resolved_by=$4, resolved_at=NOW()
		WHERE id=$1
	`, reportID, status, outcome, resolverID)
	return err
}

func (r *Repository) CreateNotification(ctx context.Context, userID int64, category, title, body string, relatedType string, relatedID *int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications(user_id,category,title,body,related_type,related_id)
		VALUES($1,$2,$3,$4,$5,$6)
	`, userID, category, title, body, relatedType, relatedID)
	return err
}

func (r *Repository) ListNotifications(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,category,title,body,related_type,related_id,read_at,created_at FROM notifications WHERE user_id=$1 ORDER BY id DESC LIMIT 200`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var category, title, body, relatedType string
		var relatedID *int64
		var readAt *time.Time
		var createdAt time.Time
		if err := rows.Scan(&id, &category, &title, &body, &relatedType, &relatedID, &readAt, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "category": category, "title": title, "body": body, "relatedType": relatedType, "relatedId": relatedID, "readAt": readAt, "createdAt": createdAt})
	}
	return items, rows.Err()
}

func (r *Repository) MarkNotificationRead(ctx context.Context, userID, notificationID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET read_at=NOW() WHERE id=$1 AND user_id=$2`, notificationID, userID)
	return err
}

func (r *Repository) QueueEmailTemplate(ctx context.Context, templateKey, recipientLabel, subject, body string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO email_template_queue(template_key,recipient_label,subject,body,status)
		VALUES($1,$2,$3,$4,'queued')
		RETURNING id
	`, templateKey, recipientLabel, subject, body).Scan(&id)
	return id, err
}

func (r *Repository) ListEmailQueue(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,template_key,recipient_label,subject,body,status,created_at,exported_at FROM email_template_queue ORDER BY id DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var templateKey, recipientLabel, subject, body, status string
		var createdAt time.Time
		var exportedAt *time.Time
		if err := rows.Scan(&id, &templateKey, &recipientLabel, &subject, &body, &status, &createdAt, &exportedAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "templateKey": templateKey, "recipientLabel": recipientLabel, "subject": subject, "body": body, "status": status, "createdAt": createdAt, "exportedAt": exportedAt})
	}
	return items, rows.Err()
}

func (r *Repository) ExportEmailQueueCSV(ctx context.Context, outDir string) (string, error) {
	items, err := r.ListEmailQueue(ctx)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("email-queue-%s.csv", time.Now().UTC().Format("20060102-150405"))
	path := filepath.Join(outDir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"id", "template_key", "recipient_label", "subject", "body", "status", "created_at"})
	for _, it := range items {
		_ = w.Write([]string{
			fmt.Sprintf("%v", it["id"]),
			fmt.Sprintf("%v", it["templateKey"]),
			fmt.Sprintf("%v", it["recipientLabel"]),
			fmt.Sprintf("%v", it["subject"]),
			fmt.Sprintf("%v", it["body"]),
			fmt.Sprintf("%v", it["status"]),
			fmt.Sprintf("%v", it["createdAt"]),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	_, _ = r.pool.Exec(ctx, `UPDATE email_template_queue SET status='exported', exported_at=NOW() WHERE status='queued'`)
	return path, nil
}

func (r *Repository) AnalyticsKPIs(ctx context.Context, from, to time.Time, providerID, packageID *int64) (map[string]any, error) {
	filters := []string{"slot_start >= $1", "slot_start < $2"}
	args := []any{from, to}
	idx := 3
	if providerID != nil {
		filters = append(filters, fmt.Sprintf("host_id = $%d", idx))
		args = append(args, *providerID)
		idx++
	}
	if packageID != nil {
		filters = append(filters, fmt.Sprintf("package_id = $%d", idx))
		args = append(args, *packageID)
	}
	where := strings.Join(filters, " AND ")

	var bookingVolume int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM bookings WHERE %s`, where), args...).Scan(&bookingVolume); err != nil {
		return nil, err
	}
	var holdCount int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM reservation_holds WHERE %s`, where), args...).Scan(&holdCount); err != nil {
		return nil, err
	}
	attendance := 0.0
	if holdCount > 0 {
		attendance = float64(bookingVolume) / float64(holdCount) * 100
	}
	return map[string]any{
		"bookingVolume":    bookingVolume,
		"attendanceRate":   attendance,
		"repurchaseRate":   0.0,
		"refundRate":       0.0,
		"coachUtilization": attendance,
	}, nil
}

func (r *Repository) ExportAnalyticsCSV(ctx context.Context, outDir string, jobID int64, reportType string, kpis map[string]any) (string, error) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}
	label := sanitizeFileLabel(reportType)
	prefix := "manual"
	if jobID > 0 {
		prefix = fmt.Sprintf("job-%d", jobID)
	}
	name := fmt.Sprintf("%s-%s-%s.csv", label, prefix, time.Now().UTC().Format("20060102-150405"))
	path := filepath.Join(outDir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"metric", "value"})
	for _, k := range []string{"bookingVolume", "attendanceRate", "repurchaseRate", "refundRate", "coachUtilization"} {
		_ = w.Write([]string{k, fmt.Sprintf("%v", kpis[k])})
	}
	w.Flush()
	return path, w.Error()
}

func (r *Repository) ScheduleReportJob(ctx context.Context, reportType string, params string, requestedBy int64, when time.Time) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO report_jobs(report_type,parameters,status,requested_by,scheduled_for)
		VALUES($1, $2::jsonb, 'scheduled', $3, $4)
		RETURNING id
	`, reportType, params, requestedBy, when).Scan(&id)
	return id, err
}

func (r *Repository) ListReportJobs(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, report_type, parameters, status, output_path, requested_by, scheduled_for, created_at, completed_at
		FROM report_jobs
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var reportType, status, outputPath string
		var parameters []byte
		var requestedBy *int64
		var scheduledFor, createdAt time.Time
		var completedAt *time.Time
		if err := rows.Scan(&id, &reportType, &parameters, &status, &outputPath, &requestedBy, &scheduledFor, &createdAt, &completedAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id":           id,
			"reportType":   reportType,
			"parameters":   string(parameters),
			"status":       status,
			"outputPath":   outputPath,
			"requestedBy":  requestedBy,
			"scheduledFor": scheduledFor,
			"createdAt":    createdAt,
			"completedAt":  completedAt,
		})
	}
	return items, rows.Err()
}

func (r *Repository) ProcessDueReportJobs(ctx context.Context, outDir string) error {
	rows, err := r.pool.Query(ctx, `
		SELECT id, report_type, parameters, scheduled_for
		FROM report_jobs
		WHERE status='scheduled' AND scheduled_for <= NOW()
		ORDER BY scheduled_for ASC
		LIMIT 50
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	jobs := []struct {
		id           int64
		reportType   string
		parameters   []byte
		scheduledFor time.Time
	}{}
	for rows.Next() {
		var job struct {
			id           int64
			reportType   string
			parameters   []byte
			scheduledFor time.Time
		}
		if err := rows.Scan(&job.id, &job.reportType, &job.parameters, &job.scheduledFor); err != nil {
			return err
		}
		jobs = append(jobs, job)
	}
	for _, job := range jobs {
		from, to, providerID, packageID, paramsErr := parseReportJobParams(job.parameters, job.scheduledFor)
		if paramsErr != nil {
			_, _ = r.pool.Exec(ctx, `UPDATE report_jobs SET status='failed' WHERE id=$1`, job.id)
			continue
		}
		kpis, err := r.AnalyticsKPIs(ctx, from, to, providerID, packageID)
		if err != nil {
			_, _ = r.pool.Exec(ctx, `UPDATE report_jobs SET status='failed' WHERE id=$1`, job.id)
			continue
		}
		path, err := r.ExportAnalyticsCSV(ctx, outDir, job.id, job.reportType, kpis)
		if err != nil {
			_, _ = r.pool.Exec(ctx, `UPDATE report_jobs SET status='failed' WHERE id=$1`, job.id)
			continue
		}
		_, _ = r.pool.Exec(ctx, `
			UPDATE report_jobs
			SET status='completed', output_path=$2, completed_at=NOW()
			WHERE id=$1
		`, job.id, path)
	}
	return nil
}

func parseReportJobParams(raw []byte, scheduledFor time.Time) (time.Time, time.Time, *int64, *int64, error) {
	from := scheduledFor.Add(-24 * time.Hour)
	to := scheduledFor
	var providerID *int64
	var packageID *int64
	if len(raw) == 0 {
		return from, to, providerID, packageID, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return time.Time{}, time.Time{}, nil, nil, err
	}
	if v, ok := payload["from"]; ok {
		if parsed, err := parseReportTime(v); err == nil {
			from = parsed
		} else {
			return time.Time{}, time.Time{}, nil, nil, err
		}
	}
	if v, ok := payload["to"]; ok {
		if parsed, err := parseReportTime(v); err == nil {
			to = parsed
		} else {
			return time.Time{}, time.Time{}, nil, nil, err
		}
	}
	if v, ok := payload["providerId"]; ok {
		if parsed, err := parseReportID(v); err == nil {
			providerID = parsed
		} else {
			return time.Time{}, time.Time{}, nil, nil, err
		}
	}
	if v, ok := payload["packageId"]; ok {
		if parsed, err := parseReportID(v); err == nil {
			packageID = parsed
		} else {
			return time.Time{}, time.Time{}, nil, nil, err
		}
	}
	return from, to, providerID, packageID, nil
}

func parseReportTime(value any) (time.Time, error) {
	switch v := value.(type) {
	case string:
		return time.Parse(time.RFC3339, v)
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("invalid time parameter: %v", value)
	}
}

func parseReportID(value any) (*int64, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case float64:
		out := int64(v)
		return &out, nil
	case json.Number:
		num, err := v.Int64()
		if err != nil {
			return nil, err
		}
		return &num, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		return &num, nil
	default:
		return nil, fmt.Errorf("invalid id parameter: %T", value)
	}
}

func sanitizeFileLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return "analytics"
	}
	label = strings.ToLower(label)
	label = strings.ReplaceAll(label, "_", "-")
	label = strings.ReplaceAll(label, " ", "-")
	label = strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, label)
	if label == "" {
		return "analytics"
	}
	return label
}

func (r *Repository) LikeTarget(ctx context.Context, userID int64, targetType string, targetID int64) error {
	switch targetType {
	case "post":
		_, err := r.pool.Exec(ctx, `
			INSERT INTO community_reactions(user_id, post_id, reaction_type)
			VALUES($1,$2,'like')
			ON CONFLICT (user_id, post_id, comment_id, reaction_type) DO NOTHING
		`, userID, targetID)
		return err
	case "comment":
		_, err := r.pool.Exec(ctx, `
			INSERT INTO community_reactions(user_id, comment_id, reaction_type)
			VALUES($1,$2,'like')
			ON CONFLICT (user_id, post_id, comment_id, reaction_type) DO NOTHING
		`, userID, targetID)
		return err
	default:
		return fmt.Errorf("invalid targetType")
	}
}
