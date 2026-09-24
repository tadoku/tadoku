package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open does not migrate the schema or require administrative privileges.
func Open(ctx context.Context, dsn string, maxConnections int32) (*pgxpool.Pool, error) {
	if maxConnections < 1 {
		return nil, fmt.Errorf("postgres max connections must be positive")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = maxConnections
	cfg.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
