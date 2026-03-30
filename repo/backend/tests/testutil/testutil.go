package testutil

import (
	"context"
	"os"
	"testing"

	"meridian/backend/internal/platform/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBPoolOrSkip connects to the DATABASE_URL and returns a pgx pool.
// If DATABASE_URL is not set, the test is skipped.
func DBPoolOrSkip(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping DB integration test")
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	return pool
}
