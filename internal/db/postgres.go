package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Logger interface {
	Info(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

type ReadyChecker struct {
	db *sql.DB
}

func NewReadyChecker(db *sql.DB) ReadyChecker {
	return ReadyChecker{db: db}
}

func (r ReadyChecker) Ready(ctx context.Context) error {
	readyCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return r.db.PingContext(readyCtx)
}

func OpenWithRetry(ctx context.Context, cfg Config, logger Logger, attempts int, interval time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		lastErr = db.PingContext(pingCtx)
		cancel()

		if lastErr == nil {
			logger.Info("database connected", "host", cfg.Host, "port", cfg.Port, "database", cfg.Name)
			return db, nil
		}

		logger.Error("database ping failed", "attempt", attempt, "error", lastErr.Error())
		if attempt == attempts {
			break
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			db.Close()
			return nil, ctx.Err()
		}
	}

	db.Close()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, ctx.Err()
	}

	return nil, fmt.Errorf("database ping failed after %d attempts: %w", attempts, lastErr)
}
