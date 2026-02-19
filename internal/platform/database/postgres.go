package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns          = 50
	defaultMinConns          = 10
	defaultMaxConnLifetime   = 1 * time.Hour
	defaultMaxConnIdleTime   = 30 * time.Minute
	defaultHealthCheckPeriod = 1 * time.Minute
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, databaseURL string) (*Postgres, error) {
	poolCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: unable to parse DATABASE_URL: %w", err)
	}

	poolCfg.MaxConns = defaultMaxConns
	poolCfg.MinConns = defaultMinConns
	poolCfg.MaxConnLifetime = defaultMaxConnLifetime
	poolCfg.MaxConnIdleTime = defaultMaxConnIdleTime
	poolCfg.HealthCheckPeriod = defaultHealthCheckPeriod

	// create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close() // release any resources acquired so far
		return nil, fmt.Errorf("database: ping failed — is PostgreSQL running? %w", err)
	}

	return &Postgres{Pool: pool}, nil
}

// gracefully shuts down the connection pool
func (pg *Postgres) Close() {
	if pg.Pool != nil {
		pg.Pool.Close()
	}
}
