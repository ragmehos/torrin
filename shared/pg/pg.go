package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, dsn, schema string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
