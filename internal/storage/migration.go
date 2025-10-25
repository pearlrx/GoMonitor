package storage

import (
	"context"
	"fmt"
	"os"
)

// RunMigrations Function for migration tables

func (pg *Postgres) RunMigrations(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	_, err = pg.pool.Exec(ctx, string(data))
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	fmt.Println("[migrate] migrations applied successfully ✅")
	return nil
}
