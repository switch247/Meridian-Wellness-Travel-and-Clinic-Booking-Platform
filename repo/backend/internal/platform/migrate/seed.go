package migrate

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Seed(ctx context.Context, pool *pgxpool.Pool, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}
	if _, err := pool.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("exec seed: %w", err)
	}
	return nil
}
