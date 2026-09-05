package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection struct {
	pool *pgxpool.Pool
}

func NewConnection(ctx context.Context, dsn string) (*Connection, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Connection{pool: pool}, nil
}

func (c *Connection) Close() {
	c.pool.Close()
}
