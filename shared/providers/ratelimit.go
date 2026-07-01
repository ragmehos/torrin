package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type limitSpec struct {
	max      int
	interval time.Duration
}

type rlDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type keyLimiter struct {
	id    string
	specs []limitSpec
	mem   []*rdLimiter
}

var (
	limMu    sync.Mutex
	limCache = map[string]*keyLimiter{}
	sharedDB rlDB
)

func SetRateLimitDB(ctx context.Context, db rlDB) error {
	limMu.Lock()
	sharedDB = db
	limMu.Unlock()
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS rate_limits (
		id TEXT PRIMARY KEY,
		tokens DOUBLE PRECISION NOT NULL,
		last_refill TIMESTAMPTZ NOT NULL DEFAULT now())`)
	return err
}

func sharedLimiters(id string, specs ...limitSpec) *keyLimiter {
	limMu.Lock()
	defer limMu.Unlock()
	if k, ok := limCache[id]; ok {
		return k
	}
	mem := make([]*rdLimiter, len(specs))
	for i, s := range specs {
		mem[i] = newRDLimiter(s.max, s.interval)
	}
	k := &keyLimiter{id: id, specs: specs, mem: mem}
	limCache[id] = k
	return k
}

func waitLimits(ctx context.Context, k *keyLimiter) {
	limMu.Lock()
	db := sharedDB
	limMu.Unlock()
	if db == nil {
		for _, m := range k.mem {
			m.wait(ctx)
		}
		return
	}
	for i, s := range k.specs {
		dbWait(ctx, db, fmt.Sprintf("%s#%d", k.id, i), float64(s.max)/s.interval.Seconds(), float64(s.max))
	}
}

func dbWait(ctx context.Context, db rlDB, id string, rate, max float64) {
	for {
		if dbTake(ctx, db, id, rate, max) {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func dbTake(ctx context.Context, db rlDB, id string, rate, max float64) bool {
	var ok bool
	err := db.QueryRow(ctx, `
		INSERT INTO rate_limits (id, tokens, last_refill) VALUES ($1, $2 - 1, now())
		ON CONFLICT (id) DO UPDATE SET
			tokens = LEAST($2, rate_limits.tokens + EXTRACT(EPOCH FROM (now() - rate_limits.last_refill)) * $3) - 1,
			last_refill = now()
		WHERE LEAST($2, rate_limits.tokens + EXTRACT(EPOCH FROM (now() - rate_limits.last_refill)) * $3) >= 1
		RETURNING true`, id, max, rate).Scan(&ok)
	if err != nil {
		return !errors.Is(err, pgx.ErrNoRows)
	}
	return ok
}
