package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"meridian/backend/tests/testutil"
)

func TestAssignRole_AdminEscalationGuard(t *testing.T) {
	pool := testutil.DBPoolOrSkip(t)
	repo := New(pool)
	ctx := context.Background()

	// unique usernames to avoid conflicts
	actorUsername := fmt.Sprintf("actor_ops_%d", time.Now().UnixNano())
	adminUsername := fmt.Sprintf("actor_admin_%d", time.Now().UnixNano())
	targetUsername := fmt.Sprintf("target_%d", time.Now().UnixNano())

	actorID, err := repo.CreateUser(ctx, actorUsername, "hash", "", "")
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	adminID, err := repo.CreateUser(ctx, adminUsername, "hash", "", "")
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	targetID, err := repo.CreateUser(ctx, targetUsername, "hash", "", "")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}

	// grant roles
	if _, err := pool.Exec(ctx, `INSERT INTO user_roles(user_id,role_name) VALUES($1,$2)`, actorID, "operations"); err != nil {
		t.Fatalf("assign operations role: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO user_roles(user_id,role_name) VALUES($1,$2)`, adminID, "admin"); err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	// operations actor must NOT be able to assign 'admin'
	if err := repo.AssignRole(ctx, actorID, targetID, "admin"); err == nil {
		t.Fatalf("expected error assigning admin by non-admin, got nil")
	}

	// admin actor should be able to assign 'admin'
	if err := repo.AssignRole(ctx, adminID, targetID, "admin"); err != nil {
		t.Fatalf("admin should assign admin: %v", err)
	}

	// verify target has admin role
	roles, err := repo.GetRolesForUser(ctx, targetID)
	if err != nil {
		t.Fatalf("get roles: %v", err)
	}
	found := false
	for _, r := range roles {
		if r == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected target to have admin role")
	}

	// cleanup (best-effort)
	pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id IN ($1,$2,$3)`, actorID, adminID, targetID)
	pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, actorID, adminID, targetID)
}
